package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Pokemon holds scraped Pokémon data
type Pokemon struct {
	NationalNo int    `json:"nationalNo"`
	Name       string `json:"name"`
	NameCN     string `json:"nameCN"`
	Type1      string `json:"type1"`
	Type2      string `json:"type2"`
	Form       string `json:"form,omitempty"` // regional variant mark
}

var typeMap = map[string]string{
	"Normal":   "Normal",
	"Fire":     "Fire",
	"Water":    "Water",
	"Electric": "Electric",
	"Grass":    "Grass",
	"Ice":      "Ice",
	"Fighting": "Fighting",
	"Poison":   "Poison",
	"Ground":   "Ground",
	"Flying":   "Flying",
	"Psychic":  "Psychic",
	"Bug":      "Bug",
	"Rock":     "Rock",
	"Ghost":    "Ghost",
	"Dragon":   "Dragon",
	"Dark":     "Dark",
	"Steel":    "Steel",
	"Fairy":    "Fairy",
	// Chinese type names (from 52poke)
	"一般": "Normal", "火": "Fire", "水": "Water", "电": "Electric",
	"草": "Grass", "冰": "Ice", "格斗": "Fighting", "毒": "Poison",
	"地面": "Ground", "飞行": "Flying", "超能力": "Psychic", "虫": "Bug",
	"岩石": "Rock", "幽灵": "Ghost", "龙": "Dragon", "恶": "Dark",
	"钢": "Steel", "妖精": "Fairy",
}

func main() {
	fmt.Println("=== Pokémon Importer ===")
	fmt.Println()

	// Try fetching from 52poke wiki
	var pokemon []Pokemon
	var err error

	fmt.Println("正在从神奇宝贝百科获取数据...")
	pokemon, err = scrape52Poke()
	if err != nil {
		fmt.Printf("⚠ 网络抓取失败: %v\n", err)
		fmt.Println("使用内置 fallback 数据...")
		pokemon = getFallbackData()
	} else {
		fmt.Printf("✅ 成功获取 %d 只宝可梦\n", len(pokemon))
	}

	// Count by type
	typeCount := map[string]int{}
	regionalCount := 0
	for _, p := range pokemon {
		typeCount[p.Type1]++
		if p.Type2 != "" {
			typeCount[p.Type2]++
		}
		if p.Form != "" {
			regionalCount++
		}
	}

	fmt.Println()
	fmt.Printf("总计: %d 只宝可梦（含 %d 种地区形态）\n", len(pokemon), regionalCount)
	fmt.Println("属性分布:")
	for _, t := range []string{"Normal", "Fire", "Water", "Electric", "Grass", "Ice", "Fighting", "Poison", "Ground", "Flying", "Psychic", "Bug", "Rock", "Ghost", "Dragon", "Dark", "Steel", "Fairy"} {
		if c, ok := typeCount[t]; ok {
			fmt.Printf("  %s: %d\n", t, c)
		}
	}

	// Save to JSON
	output := "pokemon_data.json"
	if len(os.Args) > 1 {
		output = os.Args[1]
	}

	data, _ := json.MarshalIndent(pokemon, "", "  ")
	if err := os.WriteFile(output, data, 0644); err != nil {
		fmt.Printf("❌ 写入文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n✅ 数据已保存到: %s\n", output)
	fmt.Println()
	fmt.Println("导入到数据库:")
	fmt.Printf("  go run %s/seed/import.go\n", output)
}

// scrape52Poke fetches the Pokémon list from 52poke wiki and parses it
func scrape52Poke() ([]Pokemon, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	// Use the API to get parsed page content
	apiURL := "https://wiki.52poke.com/api.php?action=parse&page=%E5%AE%9D%E5%8F%AF%E6%A2%A6%E5%88%97%E8%A1%A8%EF%BC%88%E6%8C%89%E5%85%A8%E5%9B%BD%E5%9B%BE%E9%89%B4%E7%BC%96%E5%8F%B7%EF%BC%89&format=json&prop=text&section=2"

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// Parse JSON response
	var apiResp struct {
		Parse struct {
			Text struct {
				Content string `json:"*"`
			} `json:"text"`
		} `json:"parse"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	htmlContent := apiResp.Parse.Text.Content
	if htmlContent == "" {
		return nil, fmt.Errorf("未获取到页面内容")
	}

	return parseHTMLTable(htmlContent)
}

// parseHTMLTable extracts Pokémon data from the wiki HTML table
func parseHTMLTable(htmlContent string) ([]Pokemon, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("HTML 解析失败: %w", err)
	}

	var pokemon []Pokemon
	no := 0

	// Find all tables
	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			rows := extractTableRows(n)
			for _, row := range rows {
				if len(row) < 4 {
					continue
				}

				// Skip header rows
				firstText := getCellText(row[0])
				if firstText == "" || firstText == "#" || strings.HasPrefix(firstText, "全国") {
					continue
				}

				// Extract national number
				noText := strings.TrimSpace(firstText)
				noText = strings.TrimPrefix(noText, "#")

				nationalNo, err := strconv.Atoi(noText)
				if err != nil {
					continue
				}

				// Extract name (English)
				engName := extractEnglishName(row[1])
				// Extract Chinese name
				cnName := extractChineseName(row[1])

				if engName == "" && cnName == "" {
					continue
				}

				// Extract types
				type1, type2 := extractTypes(row)

				no++
				pokemon = append(pokemon, Pokemon{
					NationalNo: nationalNo,
					Name:       engName,
					NameCN:     cnName,
					Type1:      type1,
					Type2:      type2,
				})
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTable(c)
		}
	}

	findTable(doc)

	if len(pokemon) == 0 {
		return nil, fmt.Errorf("未找到宝可梦数据")
	}

	return pokemon, nil
}

// extractTableRows returns all rows from a table
func extractTableRows(table *html.Node) [][]*html.Node {
	var rows [][]*html.Node
	var currentRow []*html.Node

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "tbody" {
				
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "tr" {
						currentRow = nil
						for td := c.FirstChild; td != nil; td = td.NextSibling {
							if td.Type == html.ElementNode && (td.Data == "td" || td.Data == "th") {
								currentRow = append(currentRow, td)
							}
						}
						rows = append(rows, currentRow)
					}
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(table)

	return rows
}

// getCellText extracts all text from a cell
func getCellText(cell *html.Node) string {
	var text string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(cell)
	return strings.TrimSpace(text)
}

// extractEnglishName tries to find the English name from a cell
func extractEnglishName(cell *html.Node) string {
	// Try to find English text in the cell
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, attr := range n.Attr {
				if attr.Key == "lang" && attr.Val == "en" {
					text := getCellText(n)
					if text != "" {
						parts = append(parts, text)
					}
					return
				}
			}
		}
		if n.Type == html.TextNode {
			t := strings.TrimSpace(n.Data)
			if t != "" && len(t) > 2 {
				parts = append(parts, t)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(cell)

	// English name is typically the longest word
	if len(parts) > 0 {
		// Filter for English-only characters
		for _, p := range parts {
			if matched, _ := regexp.MatchString(`^[A-Za-z\'\-\.]+$`, p); matched {
				return p
			}
		}
		return parts[0]
	}
	return ""
}

// extractChineseName extracts Chinese name from a cell
func extractChineseName(cell *html.Node) string {
	var text string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			// Check if it's a link to a Pokémon page
			for _, attr := range n.Attr {
				if attr.Key == "title" && strings.Contains(attr.Val, "宝可梦") {
					title := attr.Val
					if idx := strings.Index(title, "（"); idx > 0 {
						title = title[:idx]
					}
					if text == "" {
						text = title
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(cell)

	if text == "" {
		text = getCellText(cell)
	}

	return text
}

// extractTypes extracts Pokémon types from the row
func extractTypes(row []*html.Node) (string, string) {
	var types []string

	for _, cell := range row {
		cellText := getCellText(cell)

		// Check each type
		for _, t := range []string{"Normal", "Fire", "Water", "Electric", "Grass", "Ice",
			"Fighting", "Poison", "Ground", "Flying", "Psychic", "Bug",
			"Rock", "Ghost", "Dragon", "Dark", "Steel", "Fairy"} {
			if strings.Contains(cellText, t) {
				types = append(types, t)
			}
		}

		// Also check Chinese type names
		for cn, en := range typeMap {
			if strings.Contains(cellText, cn) && !contains(types, en) {
				types = append(types, en)
			}
		}

		if len(types) >= 2 {
			break
		}
	}

	if len(types) == 0 {
		return "Normal", ""
	}
	if len(types) >= 2 {
		return types[0], types[1]
	}
	return types[0], ""
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// fallbackData returns a massive hardcoded Pokémon dataset covering all 1025
func getFallbackData() []Pokemon {
	var pokemon []Pokemon

	// Helper to append with validation
	add := func(no int, name, nameCN, type1, type2 string) {
		if name == "" || nameCN == "" {
			return
		}
		pokemon = append(pokemon, Pokemon{
			NationalNo: no,
			Name:       name,
			NameCN:     nameCN,
			Type1:      type1,
			Type2:      type2,
		})
	}

	// ---------- Generation 1: Kanto #001-151 ----------
	add(1, "Bulbasaur", "妙蛙种子", "Grass", "Poison")
	add(2, "Ivysaur", "妙蛙草", "Grass", "Poison")
	add(3, "Venusaur", "妙蛙花", "Grass", "Poison")
	add(4, "Charmander", "小火龙", "Fire", "")
	add(5, "Charmeleon", "火恐龙", "Fire", "")
	add(6, "Charizard", "喷火龙", "Fire", "Flying")
	add(7, "Squirtle", "杰尼龟", "Water", "")
	add(8, "Wartortle", "卡咪龟", "Water", "")
	add(9, "Blastoise", "水箭龟", "Water", "")
	add(10, "Caterpie", "绿毛虫", "Bug", "")
	add(11, "Metapod", "铁甲蛹", "Bug", "")
	add(12, "Butterfree", "巴大蝶", "Bug", "Flying")
	add(13, "Weedle", "独角虫", "Bug", "Poison")
	add(14, "Kakuna", "铁壳蛹", "Bug", "Poison")
	add(15, "Beedrill", "大针蜂", "Bug", "Poison")
	add(16, "Pidgey", "波波", "Normal", "Flying")
	add(17, "Pidgeotto", "比比鸟", "Normal", "Flying")
	add(18, "Pidgeot", "大比鸟", "Normal", "Flying")
	add(19, "Rattata", "小拉达", "Normal", "")
	add(20, "Raticate", "拉达", "Normal", "")
	add(21, "Spearow", "烈雀", "Normal", "Flying")
	add(22, "Fearow", "大嘴雀", "Normal", "Flying")
	add(23, "Ekans", "阿柏蛇", "Poison", "")
	add(24, "Arbok", "阿柏怪", "Poison", "")
	add(25, "Pikachu", "皮卡丘", "Electric", "")
	add(26, "Raichu", "雷丘", "Electric", "")
	add(27, "Sandshrew", "穿山鼠", "Ground", "")
	add(28, "Sandslash", "穿山王", "Ground", "")
	add(29, "Nidoran♀", "尼多兰", "Poison", "")
	add(30, "Nidorina", "尼多娜", "Poison", "")
	add(31, "Nidoqueen", "尼多后", "Poison", "Ground")
	add(32, "Nidoran♂", "尼多朗", "Poison", "")
	add(33, "Nidorino", "尼多力诺", "Poison", "")
	add(34, "Nidoking", "尼多王", "Poison", "Ground")
	add(35, "Clefairy", "皮皮", "Fairy", "")
	add(36, "Clefable", "皮可西", "Fairy", "")
	add(37, "Vulpix", "六尾", "Fire", "")
	add(38, "Ninetales", "九尾", "Fire", "")
	add(39, "Jigglypuff", "胖丁", "Normal", "Fairy")
	add(40, "Wigglytuff", "胖可丁", "Normal", "Fairy")
	add(41, "Zubat", "超音蝠", "Poison", "Flying")
	add(42, "Golbat", "大嘴蝠", "Poison", "Flying")
	add(43, "Oddish", "走路草", "Grass", "Poison")
	add(44, "Gloom", "臭臭花", "Grass", "Poison")
	add(45, "Vileplume", "霸王花", "Grass", "Poison")
	add(46, "Paras", "派拉斯", "Bug", "Grass")
	add(47, "Parasect", "派拉斯特", "Bug", "Grass")
	add(48, "Venonat", "毛球", "Bug", "Poison")
	add(49, "Venomoth", "摩鲁蛾", "Bug", "Poison")
	add(50, "Diglett", "地鼠", "Ground", "")
	add(51, "Dugtrio", "三地鼠", "Ground", "")
	add(52, "Meowth", "喵喵", "Normal", "")
	add(53, "Persian", "猫老大", "Normal", "")
	add(54, "Psyduck", "可达鸭", "Water", "")
	add(55, "Golduck", "哥达鸭", "Water", "")
	add(56, "Mankey", "猴怪", "Fighting", "")
	add(57, "Primeape", "火暴猴", "Fighting", "")
	add(58, "Growlithe", "卡蒂狗", "Fire", "")
	add(59, "Arcanine", "风速狗", "Fire", "")
	add(60, "Poliwag", "蚊香蝌蚪", "Water", "")
	add(61, "Poliwhirl", "蚊香君", "Water", "")
	add(62, "Poliwrath", "蚊香泳士", "Water", "Fighting")
	add(63, "Abra", "凯西", "Psychic", "")
	add(64, "Kadabra", "勇基拉", "Psychic", "")
	add(65, "Alakazam", "胡地", "Psychic", "")
	add(66, "Machop", "腕力", "Fighting", "")
	add(67, "Machoke", "豪力", "Fighting", "")
	add(68, "Machamp", "怪力", "Fighting", "")
	add(69, "Bellsprout", "喇叭芽", "Grass", "Poison")
	add(70, "Weepinbell", "口呆花", "Grass", "Poison")
	add(71, "Victreebel", "大食花", "Grass", "Poison")
	add(72, "Tentacool", "玛瑙水母", "Water", "Poison")
	add(73, "Tentacruel", "毒刺水母", "Water", "Poison")
	add(74, "Geodude", "小拳石", "Rock", "Ground")
	add(75, "Graveler", "隆隆石", "Rock", "Ground")
	add(76, "Golem", "隆隆岩", "Rock", "Ground")
	add(77, "Ponyta", "小火马", "Fire", "")
	add(78, "Rapidash", "烈焰马", "Fire", "")
	add(79, "Slowpoke", "呆呆兽", "Water", "Psychic")
	add(80, "Slowbro", "呆壳兽", "Water", "Psychic")
	add(81, "Magnemite", "小磁怪", "Electric", "Steel")
	add(82, "Magneton", "三合一磁怪", "Electric", "Steel")
	add(83, "Farfetch'd", "大葱鸭", "Normal", "Flying")
	add(84, "Doduo", "嘟嘟", "Normal", "Flying")
	add(85, "Dodrio", "嘟嘟利", "Normal", "Flying")
	add(86, "Seel", "小海狮", "Water", "")
	add(87, "Dewgong", "白海狮", "Water", "Ice")
	add(88, "Grimer", "臭泥", "Poison", "")
	add(89, "Muk", "臭臭泥", "Poison", "")
	add(90, "Shellder", "大舌贝", "Water", "")
	add(91, "Cloyster", "刺甲贝", "Water", "Ice")
	add(92, "Gastly", "鬼斯", "Ghost", "Poison")
	add(93, "Haunter", "鬼斯通", "Ghost", "Poison")
	add(94, "Gengar", "耿鬼", "Ghost", "Poison")
	add(95, "Onix", "大岩蛇", "Rock", "Ground")
	add(96, "Drowzee", "催眠貘", "Psychic", "")
	add(97, "Hypno", "引梦貘人", "Psychic", "")
	add(98, "Krabby", "大钳蟹", "Water", "")
	add(99, "Kingler", "巨钳蟹", "Water", "")
	add(100, "Voltorb", "霹雳电球", "Electric", "")
	add(101, "Electrode", "顽皮雷弹", "Electric", "")
	add(102, "Exeggcute", "蛋蛋", "Grass", "Psychic")
	add(103, "Exeggutor", "椰蛋树", "Grass", "Psychic")
	add(104, "Cubone", "卡拉卡拉", "Ground", "")
	add(105, "Marowak", "嘎啦嘎啦", "Ground", "")
	add(106, "Hitmonlee", "飞腿郎", "Fighting", "")
	add(107, "Hitmonchan", "快拳郎", "Fighting", "")
	add(108, "Lickitung", "大舌头", "Normal", "")
	add(109, "Koffing", "瓦斯弹", "Poison", "")
	add(110, "Weezing", "双弹瓦斯", "Poison", "")
	add(111, "Rhyhorn", "独角犀牛", "Ground", "Rock")
	add(112, "Rhydon", "钻角犀兽", "Ground", "Rock")
	add(113, "Chansey", "吉利蛋", "Normal", "")
	add(114, "Tangela", "蔓藤怪", "Grass", "")
	add(115, "Kangaskhan", "袋兽", "Normal", "")
	add(116, "Horsea", "墨海马", "Water", "")
	add(117, "Seadra", "海刺龙", "Water", "")
	add(118, "Goldeen", "角金鱼", "Water", "")
	add(119, "Seaking", "金鱼王", "Water", "")
	add(120, "Staryu", "海星星", "Water", "")
	add(121, "Starmie", "宝石海星", "Water", "Psychic")
	add(122, "Mr. Mime", "魔墙人偶", "Psychic", "Fairy")
	add(123, "Scyther", "飞天螳螂", "Bug", "Flying")
	add(124, "Jynx", "迷唇姐", "Ice", "Psychic")
	add(125, "Electabuzz", "电击兽", "Electric", "")
	add(126, "Magmar", "鸭嘴火兽", "Fire", "")
	add(127, "Pinsir", "凯罗斯", "Bug", "")
	add(128, "Tauros", "肯泰罗", "Normal", "")
	add(129, "Magikarp", "鲤鱼王", "Water", "")
	add(130, "Gyarados", "暴鲤龙", "Water", "Flying")
	add(131, "Lapras", "拉普拉斯", "Water", "Ice")
	add(132, "Ditto", "百变怪", "Normal", "")
	add(133, "Eevee", "伊布", "Normal", "")
	add(134, "Vaporeon", "水伊布", "Water", "")
	add(135, "Jolteon", "雷伊布", "Electric", "")
	add(136, "Flareon", "火伊布", "Fire", "")
	add(137, "Porygon", "多边兽", "Normal", "")
	add(138, "Omanyte", "菊石兽", "Rock", "Water")
	add(139, "Omastar", "多刺菊石兽", "Rock", "Water")
	add(140, "Kabuto", "化石盔", "Rock", "Water")
	add(141, "Kabutops", "镰刀盔", "Rock", "Water")
	add(142, "Aerodactyl", "化石翼龙", "Rock", "Flying")
	add(143, "Snorlax", "卡比兽", "Normal", "")
	add(144, "Articuno", "急冻鸟", "Ice", "Flying")
	add(145, "Zapdos", "闪电鸟", "Electric", "Flying")
	add(146, "Moltres", "火焰鸟", "Fire", "Flying")
	add(147, "Dratini", "迷你龙", "Dragon", "")
	add(148, "Dragonair", "哈克龙", "Dragon", "")
	add(149, "Dragonite", "快龙", "Dragon", "Flying")
	add(150, "Mewtwo", "超梦", "Psychic", "")
	add(151, "Mew", "梦幻", "Psychic", "")

	// ---------- Generation 2: Johto #152-251 ----------
	add(152, "Chikorita", "菊草叶", "Grass", "")
	add(153, "Bayleef", "月桂叶", "Grass", "")
	add(154, "Meganium", "大竺葵", "Grass", "")
	add(155, "Cyndaquil", "火球鼠", "Fire", "")
	add(156, "Quilava", "火岩鼠", "Fire", "")
	add(157, "Typhlosion", "火暴兽", "Fire", "")
	add(158, "Totodile", "小锯鳄", "Water", "")
	add(159, "Croconaw", "蓝鳄", "Water", "")
	add(160, "Feraligatr", "大力鳄", "Water", "")
	add(161, "Sentret", "尾立", "Normal", "")
	add(162, "Furret", "大尾立", "Normal", "")
	add(163, "Hoothoot", "咕咕", "Normal", "Flying")
	add(164, "Noctowl", "猫头夜鹰", "Normal", "Flying")
	add(165, "Ledyba", "芭瓢虫", "Bug", "Flying")
	add(166, "Ledian", "安瓢虫", "Bug", "Flying")
	add(167, "Spinarak", "圆丝蛛", "Bug", "Poison")
	add(168, "Ariados", "阿利多斯", "Bug", "Poison")
	add(169, "Crobat", "叉字蝠", "Poison", "Flying")
	add(170, "Chinchou", "灯笼鱼", "Water", "Electric")
	add(171, "Lanturn", "电灯怪", "Water", "Electric")
	add(172, "Pichu", "皮丘", "Electric", "")
	add(173, "Cleffa", "皮宝宝", "Fairy", "")
	add(174, "Igglybuff", "宝宝丁", "Normal", "Fairy")
	add(175, "Togepi", "波克比", "Fairy", "")
	add(176, "Togetic", "波克基古", "Fairy", "Flying")
	add(177, "Natu", "天然雀", "Psychic", "Flying")
	add(178, "Xatu", "天然鸟", "Psychic", "Flying")
	add(179, "Mareep", "咩利羊", "Electric", "")
	add(180, "Flaaffy", "茸茸羊", "Electric", "")
	add(181, "Ampharos", "电龙", "Electric", "")
	add(182, "Bellossom", "美丽花", "Grass", "")
	add(183, "Marill", "玛力露", "Water", "Fairy")
	add(184, "Azumarill", "玛力露丽", "Water", "Fairy")
	add(185, "Sudowoodo", "树才怪", "Rock", "")
	add(186, "Politoed", "蚊香蛙皇", "Water", "")
	add(187, "Hoppip", "毽子草", "Grass", "Flying")
	add(188, "Skiploom", "毽子花", "Grass", "Flying")
	add(189, "Jumpluff", "毽子棉", "Grass", "Flying")
	add(190, "Aipom", "长尾怪手", "Normal", "")
	add(191, "Sunkern", "向日种子", "Grass", "")
	add(192, "Sunflora", "向日花怪", "Grass", "")
	add(193, "Yanma", "蜻蜻蜓", "Bug", "Flying")
	add(194, "Wooper", "乌波", "Water", "Ground")
	add(195, "Quagsire", "沼王", "Water", "Ground")
	add(196, "Espeon", "太阳伊布", "Psychic", "")
	add(197, "Umbreon", "月亮伊布", "Dark", "")
	add(198, "Murkrow", "黑暗鸦", "Dark", "Flying")
	add(199, "Slowking", "呆呆王", "Water", "Psychic")
	add(200, "Misdreavus", "梦妖", "Ghost", "")
	add(201, "Unown", "未知图腾", "Psychic", "")
	add(202, "Wobbuffet", "果然翁", "Psychic", "")
	add(203, "Girafarig", "麒麟奇", "Normal", "Psychic")
	add(204, "Pineco", "榛果球", "Bug", "")
	add(205, "Forretress", "佛烈托斯", "Bug", "Steel")
	add(206, "Dunsparce", "土龙弟弟", "Normal", "")
	add(207, "Gligar", "天蝎", "Ground", "Flying")
	add(208, "Steelix", "大钢蛇", "Steel", "Ground")
	add(209, "Snubbull", "布鲁", "Fairy", "")
	add(210, "Granbull", "布鲁皇", "Fairy", "")
	add(211, "Qwilfish", "千针鱼", "Water", "Poison")
	add(212, "Scizor", "巨钳螳螂", "Bug", "Steel")
	add(213, "Shuckle", "壶壶", "Bug", "Rock")
	add(214, "Heracross", "赫拉克罗斯", "Bug", "Fighting")
	add(215, "Sneasel", "狃拉", "Dark", "Ice")
	add(216, "Teddiursa", "熊宝宝", "Normal", "")
	add(217, "Ursaring", "圈圈熊", "Normal", "")
	add(218, "Slugma", "熔岩虫", "Fire", "")
	add(219, "Magcargo", "熔岩蜗牛", "Fire", "Rock")
	add(220, "Swinub", "小山猪", "Ice", "Ground")
	add(221, "Piloswine", "长毛猪", "Ice", "Ground")
	add(222, "Corsola", "太阳珊瑚", "Water", "Rock")
	add(223, "Remoraid", "铁炮鱼", "Water", "")
	add(224, "Octillery", "章鱼桶", "Water", "")
	add(225, "Delibird", "信使鸟", "Ice", "Flying")
	add(226, "Mantine", "巨翅飞鱼", "Water", "Flying")
	add(227, "Skarmory", "盔甲鸟", "Steel", "Flying")
	add(228, "Houndour", "戴鲁比", "Dark", "Fire")
	add(229, "Houndoom", "黑鲁加", "Dark", "Fire")
	add(230, "Kingdra", "刺龙王", "Water", "Dragon")
	add(231, "Phanpy", "小小象", "Ground", "")
	add(232, "Donphan", "顿甲", "Ground", "")
	add(233, "Porygon2", "多边兽Ⅱ", "Normal", "")
	add(234, "Stantler", "惊角鹿", "Normal", "")
	add(235, "Smeargle", "图图犬", "Normal", "")
	add(236, "Tyrogue", "无畏小子", "Fighting", "")
	add(237, "Hitmontop", "战舞郎", "Fighting", "")
	add(238, "Smoochum", "迷唇娃", "Ice", "Psychic")
	add(239, "Elekid", "电击怪", "Electric", "")
	add(240, "Magby", "鸭嘴宝宝", "Fire", "")
	add(241, "Miltank", "大奶罐", "Normal", "")
	add(242, "Blissey", "幸福蛋", "Normal", "")
	add(243, "Raikou", "雷公", "Electric", "")
	add(244, "Entei", "炎帝", "Fire", "")
	add(245, "Suicune", "水君", "Water", "")
	add(246, "Larvitar", "幼基拉斯", "Rock", "Ground")
	add(247, "Pupitar", "沙基拉斯", "Rock", "Ground")
	add(248, "Tyranitar", "班基拉斯", "Rock", "Dark")
	add(249, "Lugia", "洛奇亚", "Psychic", "Flying")
	add(250, "Ho-Oh", "凤王", "Fire", "Flying")
	add(251, "Celebi", "时拉比", "Psychic", "Grass")

	// ---------- Generation 3: Hoenn #252-386 ----------
	add(252, "Treecko", "木守宫", "Grass", "")
	add(253, "Grovyle", "森林蜥蜴", "Grass", "")
	add(254, "Sceptile", "蜥蜴王", "Grass", "")
	add(255, "Torchic", "火稚鸡", "Fire", "")
	add(256, "Combusken", "力壮鸡", "Fire", "Fighting")
	add(257, "Blaziken", "火焰鸡", "Fire", "Fighting")
	add(258, "Mudkip", "水跃鱼", "Water", "")
	add(259, "Marshtomp", "沼跃鱼", "Water", "Ground")
	add(260, "Swampert", "巨沼怪", "Water", "Ground")
	add(261, "Poochyena", "土狼犬", "Dark", "")
	add(262, "Mightyena", "大狼犬", "Dark", "")
	add(263, "Zigzagoon", "蛇纹熊", "Normal", "")
	add(264, "Linoone", "直冲熊", "Normal", "")
	add(265, "Wurmple", "刺尾虫", "Bug", "")
	add(266, "Silcoon", "甲壳茧", "Bug", "")
	add(267, "Beautifly", "狩猎凤蝶", "Bug", "Flying")
	add(268, "Cascoon", "盾甲茧", "Bug", "")
	add(269, "Dustox", "毒粉蛾", "Bug", "Poison")
	add(270, "Lotad", "莲叶童子", "Water", "Grass")
	add(271, "Lombre", "莲帽小童", "Water", "Grass")
	add(272, "Ludicolo", "乐天河童", "Water", "Grass")
	add(273, "Seedot", "橡实果", "Grass", "")
	add(274, "Nuzleaf", "长鼻叶", "Grass", "Dark")
	add(275, "Shiftry", "狡猾天狗", "Grass", "Dark")
	add(276, "Taillow", "傲骨燕", "Normal", "Flying")
	add(277, "Swellow", "大王燕", "Normal", "Flying")
	add(278, "Wingull", "长翅鸥", "Water", "Flying")
	add(279, "Pelipper", "大嘴鸥", "Water", "Flying")
	add(280, "Ralts", "拉鲁拉丝", "Psychic", "Fairy")
	add(281, "Kirlia", "奇鲁莉安", "Psychic", "Fairy")
	add(282, "Gardevoir", "沙奈朵", "Psychic", "Fairy")
	add(283, "Surskit", "溜溜糖球", "Bug", "Water")
	add(284, "Masquerain", "雨翅蛾", "Bug", "Flying")
	add(285, "Shroomish", "蘑蘑菇", "Grass", "")
	add(286, "Breloom", "斗笠菇", "Grass", "Fighting")
	add(287, "Slakoth", "懒人獭", "Normal", "")
	add(288, "Vigoroth", "过动猿", "Normal", "")
	add(289, "Slaking", "请假王", "Normal", "")
	add(290, "Nincada", "土居忍士", "Bug", "Ground")
	add(291, "Ninjask", "铁面忍者", "Bug", "Flying")
	add(292, "Shedinja", "脱壳忍者", "Bug", "Ghost")
	add(293, "Whismur", "咕妞妞", "Normal", "")
	add(294, "Loudred", "吼爆弹", "Normal", "")
	add(295, "Exploud", "爆音怪", "Normal", "")
	add(296, "Makuhita", "幕下力士", "Fighting", "")
	add(297, "Hariyama", "铁掌力士", "Fighting", "")
	add(298, "Azurill", "露力丽", "Normal", "Fairy")
	add(299, "Nosepass", "朝北鼻", "Rock", "")
	add(300, "Skitty", "向尾喵", "Normal", "")
	add(301, "Delcatty", "优雅猫", "Normal", "")
	add(302, "Sableye", "勾魂眼", "Dark", "Ghost")
	add(303, "Mawile", "大嘴娃", "Steel", "Fairy")
	add(304, "Aron", "可可多拉", "Steel", "Rock")
	add(305, "Lairon", "可多拉", "Steel", "Rock")
	add(306, "Aggron", "波士可多拉", "Steel", "Rock")
	add(307, "Meditite", "玛沙那", "Fighting", "Psychic")
	add(308, "Medicham", "恰雷姆", "Fighting", "Psychic")
	add(309, "Electrike", "落雷兽", "Electric", "")
	add(310, "Manectric", "雷电兽", "Electric", "")
	add(311, "Plusle", "正电拍拍", "Electric", "")
	add(312, "Minun", "负电拍拍", "Electric", "")
	add(313, "Volbeat", "电萤虫", "Bug", "")
	add(314, "Illumise", "甜甜萤", "Bug", "")
	add(315, "Roselia", "毒蔷薇", "Grass", "Poison")
	add(316, "Gulpin", "溶食兽", "Poison", "")
	add(317, "Swalot", "吞食兽", "Poison", "")
	add(318, "Carvanha", "利牙鱼", "Water", "Dark")
	add(319, "Sharpedo", "巨牙鲨", "Water", "Dark")
	add(320, "Wailmer", "吼吼鲸", "Water", "")
	add(321, "Wailord", "吼鲸王", "Water", "")
	add(322, "Numel", "呆火驼", "Fire", "Ground")
	add(323, "Camerupt", "喷火驼", "Fire", "Ground")
	add(324, "Torkoal", "煤炭龟", "Fire", "")
	add(325, "Spoink", "跳跳猪", "Psychic", "")
	add(326, "Grumpig", "噗噗猪", "Psychic", "")
	add(327, "Spinda", "晃晃斑", "Normal", "")
	add(328, "Trapinch", "大颚蚁", "Ground", "")
	add(329, "Vibrava", "超音波幼虫", "Ground", "Dragon")
	add(330, "Flygon", "沙漠蜻蜓", "Ground", "Dragon")
	add(331, "Cacnea", "刺球仙人掌", "Grass", "")
	add(332, "Cacturne", "梦歌仙人掌", "Grass", "Dark")
	add(333, "Swablu", "青绵鸟", "Normal", "Flying")
	add(334, "Altaria", "七夕青鸟", "Dragon", "Flying")
	add(335, "Zangoose", "猫鼬斩", "Normal", "")
	add(336, "Seviper", "饭匙蛇", "Poison", "")
	add(337, "Lunatone", "月石", "Rock", "Psychic")
	add(338, "Solrock", "太阳岩", "Rock", "Psychic")
	add(339, "Barboach", "泥泥鳅", "Water", "Ground")
	add(340, "Whiscash", "鲶鱼王", "Water", "Ground")
	add(341, "Corphish", "龙虾小兵", "Water", "")
	add(342, "Crawdaunt", "铁螯龙虾", "Water", "Dark")
	add(343, "Baltoy", "天秤偶", "Ground", "Psychic")
	add(344, "Claydol", "念力土偶", "Ground", "Psychic")
	add(345, "Lileep", "触手百合", "Rock", "Grass")
	add(346, "Cradily", "摇篮百合", "Rock", "Grass")
	add(347, "Anorith", "太古羽虫", "Rock", "Bug")
	add(348, "Armaldo", "太古盔甲", "Rock", "Bug")
	add(349, "Feebas", "丑丑鱼", "Water", "")
	add(350, "Milotic", "美纳斯", "Water", "")
	add(351, "Castform", "飘浮泡泡", "Normal", "")
	add(352, "Kecleon", "变隐龙", "Normal", "")
	add(353, "Shuppet", "怨影娃娃", "Ghost", "")
	add(354, "Banette", "诅咒娃娃", "Ghost", "")
	add(355, "Duskull", "夜巡灵", "Ghost", "")
	add(356, "Dusclops", "彷徨夜灵", "Ghost", "")
	add(357, "Tropius", "热带龙", "Grass", "Flying")
	add(358, "Chimecho", "风铃铃", "Psychic", "")
	add(359, "Absol", "阿勃梭鲁", "Dark", "")
	add(360, "Wynaut", "小果然", "Psychic", "")
	add(361, "Snorunt", "雪童子", "Ice", "")
	add(362, "Glalie", "冰鬼护", "Ice", "")
	add(363, "Spheal", "海豹球", "Ice", "Water")
	add(364, "Sealeo", "海魔狮", "Ice", "Water")
	add(365, "Walrein", "帝牙海狮", "Ice", "Water")
	add(366, "Clamperl", "珍珠贝", "Water", "")
	add(367, "Huntail", "猎斑鱼", "Water", "")
	add(368, "Gorebyss", "樱花鱼", "Water", "")
	add(369, "Relicanth", "古空棘鱼", "Water", "Rock")
	add(370, "Luvdisc", "爱心鱼", "Water", "")
	add(371, "Bagon", "宝贝龙", "Dragon", "")
	add(372, "Shelgon", "甲壳龙", "Dragon", "")
	add(373, "Salamence", "暴飞龙", "Dragon", "Flying")
	add(374, "Beldum", "铁哑铃", "Steel", "Psychic")
	add(375, "Metang", "金属怪", "Steel", "Psychic")
	add(376, "Metagross", "巨金怪", "Steel", "Psychic")
	add(377, "Regirock", "雷吉洛克", "Rock", "")
	add(378, "Regice", "雷吉艾斯", "Ice", "")
	add(379, "Registeel", "雷吉斯奇鲁", "Steel", "")
	add(380, "Latias", "拉帝亚斯", "Dragon", "Psychic")
	add(381, "Latios", "拉帝欧斯", "Dragon", "Psychic")
	add(382, "Kyogre", "盖欧卡", "Water", "")
	add(383, "Groudon", "固拉多", "Ground", "")
	add(384, "Rayquaza", "烈空坐", "Dragon", "Flying")
	add(385, "Jirachi", "基拉祈", "Steel", "Psychic")
	add(386, "Deoxys", "代欧奇希斯", "Psychic", "")

	// ---------- Generation 4: Sinnoh #387-493 ----------
	add(387, "Turtwig", "草苗龟", "Grass", "")
	add(388, "Grotle", "树林龟", "Grass", "")
	add(389, "Torterra", "土台龟", "Grass", "Ground")
	add(390, "Chimchar", "小火焰猴", "Fire", "")
	add(391, "Monferno", "猛火猴", "Fire", "Fighting")
	add(392, "Infernape", "烈焰猴", "Fire", "Fighting")
	add(393, "Piplup", "波加曼", "Water", "")
	add(394, "Prinplup", "波皇子", "Water", "")
	add(395, "Empoleon", "帝王拿波", "Water", "Steel")
	add(396, "Starly", "姆克儿", "Normal", "Flying")
	add(397, "Staravia", "姆克鸟", "Normal", "Flying")
	add(398, "Staraptor", "姆克鹰", "Normal", "Flying")
	add(399, "Bidoof", "大牙狸", "Normal", "")
	add(400, "Bibarel", "大尾狸", "Normal", "Water")
	add(401, "Kricketot", "圆法师", "Bug", "")
	add(402, "Kricketune", "音箱蟀", "Bug", "")
	add(403, "Shinx", "小猫怪", "Electric", "")
	add(404, "Luxio", "勒克猫", "Electric", "")
	add(405, "Luxray", "伦琴猫", "Electric", "")
	add(406, "Budew", "含羞苞", "Grass", "Poison")
	add(407, "Roserade", "罗丝雷朵", "Grass", "Poison")
	add(408, "Cranidos", "头盖龙", "Rock", "")
	add(409, "Rampardos", "战槌龙", "Rock", "")
	add(410, "Shieldon", "盾甲龙", "Rock", "Steel")
	add(411, "Bastiodon", "护城龙", "Rock", "Steel")
	add(412, "Burmy", "结草儿", "Bug", "")
	add(413, "Wormadam", "结草贵妇", "Bug", "Grass")
	add(414, "Mothim", "绅士蛾", "Bug", "Flying")
	add(415, "Combee", "三蜜蜂", "Bug", "Flying")
	add(416, "Vespiquen", "蜂女王", "Bug", "Flying")
	add(417, "Pachirisu", "帕奇利兹", "Electric", "")
	add(418, "Buizel", "泳圈鼬", "Water", "")
	add(419, "Floatzel", "浮潜鼬", "Water", "")
	add(420, "Cherubi", "樱花宝", "Grass", "")
	add(421, "Cherrim", "樱花儿", "Grass", "")
	add(422, "Shellos", "无壳海兔", "Water", "")
	add(423, "Gastrodon", "海兔兽", "Water", "Ground")
	add(424, "Ambipom", "双尾怪手", "Normal", "")
	add(425, "Drifloon", "飘飘球", "Ghost", "Flying")
	add(426, "Drifblim", "随风球", "Ghost", "Flying")
	add(427, "Buneary", "卷卷耳", "Normal", "")
	add(428, "Lopunny", "长耳兔", "Normal", "")
	add(429, "Mismagius", "梦妖魔", "Ghost", "")
	add(430, "Honchkrow", "乌鸦头头", "Dark", "Flying")
	add(431, "Glameow", "魅力喵", "Normal", "")
	add(432, "Purugly", "东施喵", "Normal", "")
	add(433, "Chingling", "铃铛响", "Psychic", "")
	add(434, "Stunky", "臭鼬噗", "Poison", "Dark")
	add(435, "Skuntank", "坦克臭鼬", "Poison", "Dark")
	add(436, "Bronzor", "铜镜怪", "Steel", "Psychic")
	add(437, "Bronzong", "青铜钟", "Steel", "Psychic")
	add(438, "Bonsly", "盆才怪", "Rock", "")
	add(439, "Mime Jr.", "魔尼尼", "Psychic", "Fairy")
	add(440, "Happiny", "小福蛋", "Normal", "")
	add(441, "Chatot", "聒噪鸟", "Normal", "Flying")
	add(442, "Spiritomb", "花岩怪", "Ghost", "Dark")
	add(443, "Gible", "圆陆鲨", "Dragon", "Ground")
	add(444, "Gabite", "尖牙陆鲨", "Dragon", "Ground")
	add(445, "Garchomp", "烈咬陆鲨", "Dragon", "Ground")
	add(446, "Munchlax", "小卡比兽", "Normal", "")
	add(447, "Riolu", "利欧路", "Fighting", "")
	add(448, "Lucario", "路卡利欧", "Fighting", "Steel")
	add(449, "Hippopotas", "沙河马", "Ground", "")
	add(450, "Hippowdon", "河马兽", "Ground", "")
	add(451, "Skorupi", "钳尾蝎", "Poison", "Bug")
	add(452, "Drapion", "龙王蝎", "Poison", "Dark")
	add(453, "Croagunk", "不良蛙", "Poison", "Fighting")
	add(454, "Toxicroak", "毒骷蛙", "Poison", "Fighting")
	add(455, "Carnivine", "尖牙笼", "Grass", "")
	add(456, "Finneon", "荧光鱼", "Water", "")
	add(457, "Lumineon", "霓虹鱼", "Water", "")
	add(458, "Mantyke", "小球飞鱼", "Water", "Flying")
	add(459, "Snover", "雪笠怪", "Grass", "Ice")
	add(460, "Abomasnow", "暴雪王", "Grass", "Ice")
	add(461, "Weavile", "玛狃拉", "Dark", "Ice")
	add(462, "Magnezone", "自爆磁怪", "Electric", "Steel")
	add(463, "Lickilicky", "大舌舔", "Normal", "")
	add(464, "Rhyperior", "超甲狂犀", "Ground", "Rock")
	add(465, "Tangrowth", "巨蔓藤", "Grass", "")
	add(466, "Electivire", "电击魔兽", "Electric", "")
	add(467, "Magmortar", "鸭嘴炎兽", "Fire", "")
	add(468, "Togekiss", "波克基斯", "Fairy", "Flying")
	add(469, "Yanmega", "远古巨蜓", "Bug", "Flying")
	add(470, "Leafeon", "叶伊布", "Grass", "")
	add(471, "Glaceon", "冰伊布", "Ice", "")
	add(472, "Gliscor", "天蝎王", "Ground", "Flying")
	add(473, "Mamoswine", "象牙猪", "Ice", "Ground")
	add(474, "Porygon-Z", "多边兽Ｚ", "Normal", "")
	add(475, "Gallade", "艾路雷朵", "Psychic", "Fighting")
	add(476, "Probopass", "大朝北鼻", "Rock", "Steel")
	add(477, "Dusknoir", "黑夜魔灵", "Ghost", "")
	add(478, "Froslass", "雪妖女", "Ice", "Ghost")
	add(479, "Rotom", "洛托姆", "Electric", "Ghost")
	add(480, "Uxie", "由克希", "Psychic", "")
	add(481, "Mesprit", "艾姆利多", "Psychic", "")
	add(482, "Azelf", "亚克诺姆", "Psychic", "")
	add(483, "Dialga", "帝牙卢卡", "Steel", "Dragon")
	add(484, "Palkia", "帕路奇亚", "Water", "Dragon")
	add(485, "Heatran", "席多蓝恩", "Fire", "Steel")
	add(486, "Regigigas", "雷吉奇卡斯", "Normal", "")
	add(487, "Giratina", "骑拉帝纳", "Ghost", "Dragon")
	add(488, "Cresselia", "克雷色利亚", "Psychic", "")
	add(489, "Phione", "霏欧纳", "Water", "")
	add(490, "Manaphy", "玛纳霏", "Water", "")
	add(491, "Darkrai", "达克莱伊", "Dark", "")
	add(492, "Shaymin", "谢米", "Grass", "")
	add(493, "Arceus", "阿尔宙斯", "Normal", "")

	// ---------- Generation 5: Unova #494-649 ----------
	add(494, "Victini", "比克提尼", "Psychic", "Fire")
	add(495, "Snivy", "藤藤蛇", "Grass", "")
	add(496, "Servine", "青藤蛇", "Grass", "")
	add(497, "Serperior", "君主蛇", "Grass", "")
	add(498, "Tepig", "暖暖猪", "Fire", "")
	add(499, "Pignite", "炒炒猪", "Fire", "Fighting")
	add(500, "Emboar", "炎武王", "Fire", "Fighting")
	add(501, "Oshawott", "水水獭", "Water", "")
	add(502, "Dewott", "双刃丸", "Water", "")
	add(503, "Samurott", "大剑鬼", "Water", "")
	add(504, "Patrat", "探探鼠", "Normal", "")
	add(505, "Watchog", "步哨鼠", "Normal", "")
	add(506, "Lillipup", "小约克", "Normal", "")
	add(507, "Herdier", "哈约克", "Normal", "")
	add(508, "Stoutland", "长毛狗", "Normal", "")
	add(509, "Purrloin", "扒手猫", "Dark", "")
	add(510, "Liepard", "酷豹", "Dark", "")
	add(511, "Pansage", "花椰猴", "Grass", "")
	add(512, "Simisage", "花椰猿", "Grass", "")
	add(513, "Pansear", "爆香猴", "Fire", "")
	add(514, "Simisear", "爆香猿", "Fire", "")
	add(515, "Panpour", "冷水猴", "Water", "")
	add(516, "Simipour", "冷水猿", "Water", "")
	add(517, "Munna", "食梦梦", "Psychic", "")
	add(518, "Musharna", "梦梦蚀", "Psychic", "")
	add(519, "Pidove", "豆豆鸽", "Normal", "Flying")
	add(520, "Tranquill", "咕咕鸽", "Normal", "Flying")
	add(521, "Unfezant", "高傲雉鸡", "Normal", "Flying")
	add(522, "Blitzle", "斑斑马", "Electric", "")
	add(523, "Zebstrika", "雷电斑马", "Electric", "")
	add(524, "Roggenrola", "石丸子", "Rock", "")
	add(525, "Boldore", "地幔岩", "Rock", "")
	add(526, "Gigalith", "庞岩怪", "Rock", "")
	add(527, "Woobat", "滚滚蝙蝠", "Psychic", "Flying")
	add(528, "Swoobat", "心蝙蝠", "Psychic", "Flying")
	add(529, "Drilbur", "螺钉地鼠", "Ground", "")
	add(530, "Excadrill", "龙头地鼠", "Ground", "Steel")
	add(531, "Audino", "差不多娃娃", "Normal", "")
	add(532, "Timburr", "搬运小匠", "Fighting", "")
	add(533, "Gurdurr", "铁骨土人", "Fighting", "")
	add(534, "Conkeldurr", "修建老匠", "Fighting", "")
	add(535, "Tympole", "圆蝌蚪", "Water", "")
	add(536, "Palpitoad", "蓝蟾蜍", "Water", "Ground")
	add(537, "Seismitoad", "蟾蜍王", "Water", "Ground")
	add(538, "Throh", "投摔鬼", "Fighting", "")
	add(539, "Sawk", "打击鬼", "Fighting", "")
	add(540, "Sewaddle", "虫宝包", "Bug", "Grass")
	add(541, "Swadloon", "宝包茧", "Bug", "Grass")
	add(542, "Leavanny", "保姆虫", "Bug", "Grass")
	add(543, "Venipede", "百足蜈蚣", "Bug", "Poison")
	add(544, "Whirlipede", "车轮球", "Bug", "Poison")
	add(545, "Scolipede", "蜈蚣王", "Bug", "Poison")
	add(546, "Cottonee", "木棉球", "Grass", "Fairy")
	add(547, "Whimsicott", "风妖精", "Grass", "Fairy")
	add(548, "Petilil", "百合根娃娃", "Grass", "")
	add(549, "Lilligant", "裙儿小姐", "Grass", "")
	add(550, "Basculin", "野蛮鲈鱼", "Water", "")
	add(551, "Sandile", "黑眼鳄", "Ground", "Dark")
	add(552, "Krokorok", "混混鳄", "Ground", "Dark")
	add(553, "Krookodile", "流氓鳄", "Ground", "Dark")
	add(554, "Darumaka", "火红不倒翁", "Fire", "")
	add(555, "Darmanitan", "达摩狒狒", "Fire", "")
	add(556, "Maractus", "沙铃仙人掌", "Grass", "")
	add(557, "Dwebble", "石居蟹", "Bug", "Rock")
	add(558, "Crustle", "岩殿居蟹", "Bug", "Rock")
	add(559, "Scraggy", "滑滑小子", "Dark", "Fighting")
	add(560, "Scrafty", "头巾混混", "Dark", "Fighting")
	add(561, "Sigilyph", "象征鸟", "Psychic", "Flying")
	add(562, "Yamask", "哭哭面具", "Ghost", "")
	add(563, "Cofagrigus", "死神棺", "Ghost", "")
	add(564, "Tirtouga", "原盖海龟", "Water", "Rock")
	add(565, "Carracosta", "肋骨海龟", "Water", "Rock")
	add(566, "Archen", "始祖小鸟", "Rock", "Flying")
	add(567, "Archeops", "始祖大鸟", "Rock", "Flying")
	add(568, "Trubbish", "破破袋", "Poison", "")
	add(569, "Garbodor", "灰尘山", "Poison", "")
	add(570, "Zorua", "索罗亚", "Dark", "")
	add(571, "Zoroark", "索罗亚克", "Dark", "")
	add(572, "Minccino", "泡沫栗鼠", "Normal", "")
	add(573, "Cinccino", "奇诺栗鼠", "Normal", "")
	add(574, "Gothita", "哥德宝宝", "Psychic", "")
	add(575, "Gothorita", "哥德小童", "Psychic", "")
	add(576, "Gothitelle", "哥德小姐", "Psychic", "")
	add(577, "Solosis", "单卵细胞球", "Psychic", "")
	add(578, "Duosion", "双卵细胞球", "Psychic", "")
	add(579, "Reuniclus", "人造细胞卵", "Psychic", "")
	add(580, "Ducklett", "鸭宝宝", "Water", "Flying")
	add(581, "Swanna", "舞天鹅", "Water", "Flying")
	add(582, "Vanillite", "迷你冰", "Ice", "")
	add(583, "Vanillish", "多多冰", "Ice", "")
	add(584, "Vanilluxe", "双倍多多冰", "Ice", "")
	add(585, "Deerling", "四季鹿", "Normal", "Grass")
	add(586, "Sawsbuck", "萌芽鹿", "Normal", "Grass")
	add(587, "Emolga", "电飞鼠", "Electric", "Flying")
	add(588, "Karrablast", "盖盖虫", "Bug", "")
	add(589, "Escavalier", "骑士蜗牛", "Bug", "Steel")
	add(590, "Foongus", "哎呀球菇", "Grass", "Poison")
	add(591, "Amoonguss", "败露球菇", "Grass", "Poison")
	add(592, "Frillish", "轻飘飘", "Water", "Ghost")
	add(593, "Jellicent", "胖嘟嘟", "Water", "Ghost")
	add(594, "Alomomola", "保母曼波", "Water", "")
	add(595, "Joltik", "电电虫", "Bug", "Electric")
	add(596, "Galvantula", "电蜘蛛", "Bug", "Electric")
	add(597, "Ferroseed", "种子铁球", "Grass", "Steel")
	add(598, "Ferrothorn", "坚果哑铃", "Grass", "Steel")
	add(599, "Klink", "齿轮儿", "Steel", "")
	add(600, "Klang", "齿轮组", "Steel", "")
	add(601, "Klinklang", "齿轮怪", "Steel", "")
	add(602, "Tynamo", "麻麻小鱼", "Electric", "")
	add(603, "Eelektrik", "麻麻鳗", "Electric", "")
	add(604, "Eelektross", "麻麻鳗鱼王", "Electric", "")
	add(605, "Elgyem", "小灰怪", "Psychic", "")
	add(606, "Beheeyem", "大宇怪", "Psychic", "")
	add(607, "Litwick", "烛光灵", "Ghost", "Fire")
	add(608, "Lampent", "灯火幽灵", "Ghost", "Fire")
	add(609, "Chandelure", "水晶灯火灵", "Ghost", "Fire")
	add(610, "Axew", "牙牙", "Dragon", "")
	add(611, "Fraxure", "斧牙龙", "Dragon", "")
	add(612, "Haxorus", "双斧战龙", "Dragon", "")
	add(613, "Cubchoo", "喷嚏熊", "Ice", "")
	add(614, "Beartic", "冻原熊", "Ice", "")
	add(615, "Cryogonal", "几何雪花", "Ice", "")
	add(616, "Shelmet", "小嘴蜗", "Bug", "")
	add(617, "Accelgor", "敏捷虫", "Bug", "")
	add(618, "Stunfisk", "泥巴鱼", "Ground", "Electric")
	add(619, "Mienfoo", "功夫鼬", "Fighting", "")
	add(620, "Mienshao", "师父鼬", "Fighting", "")
	add(621, "Druddigon", "赤面龙", "Dragon", "")
	add(622, "Golett", "泥偶小人", "Ground", "Ghost")
	add(623, "Golurk", "泥偶巨人", "Ground", "Ghost")
	add(624, "Pawniard", "驹刀小兵", "Dark", "Steel")
	add(625, "Bisharp", "劈斩司令", "Dark", "Steel")
	add(626, "Bouffalant", "爆炸头水牛", "Normal", "")
	add(627, "Rufflet", "毛头小鹰", "Normal", "Flying")
	add(628, "Braviary", "勇士雄鹰", "Normal", "Flying")
	add(629, "Vullaby", "秃鹰丫头", "Dark", "Flying")
	add(630, "Mandibuzz", "秃鹰娜", "Dark", "Flying")
	add(631, "Heatmor", "熔蚁兽", "Fire", "")
	add(632, "Durant", "铁蚁", "Bug", "Steel")
	add(633, "Deino", "单首龙", "Dark", "Dragon")
	add(634, "Zweilous", "双首暴龙", "Dark", "Dragon")
	add(635, "Hydreigon", "三首恶龙", "Dark", "Dragon")
	add(636, "Larvesta", "燃烧虫", "Bug", "Fire")
	add(637, "Volcarona", "火神蛾", "Bug", "Fire")
	add(638, "Cobalion", "勾帕路翁", "Steel", "Fighting")
	add(639, "Terrakion", "代拉基翁", "Rock", "Fighting")
	add(640, "Virizion", "毕力吉翁", "Grass", "Fighting")
	add(641, "Tornadus", "龙卷云", "Flying", "")
	add(642, "Thundurus", "雷电云", "Electric", "Flying")
	add(643, "Reshiram", "莱希拉姆", "Dragon", "Fire")
	add(644, "Zekrom", "捷克罗姆", "Dragon", "Electric")
	add(645, "Landorus", "土地云", "Ground", "Flying")
	add(646, "Kyurem", "酋雷姆", "Dragon", "Ice")
	add(647, "Keldeo", "凯路迪欧", "Water", "Fighting")
	add(648, "Meloetta", "美洛耶塔", "Normal", "Psychic")
	add(649, "Genesect", "盖诺赛克特", "Bug", "Steel")

	// ---------- Generation 6: Kalos #650-721 ----------
	add(650, "Chespin", "哈力栗", "Grass", "")
	add(651, "Quilladin", "胖胖哈力", "Grass", "")
	add(652, "Chesnaught", "布里卡隆", "Grass", "Fighting")
	add(653, "Fennekin", "火狐狸", "Fire", "")
	add(654, "Braixen", "长尾火狐", "Fire", "")
	add(655, "Delphox", "妖火红狐", "Fire", "Psychic")
	add(656, "Froakie", "呱呱泡蛙", "Water", "")
	add(657, "Frogadier", "呱头蛙", "Water", "")
	add(658, "Greninja", "甲贺忍蛙", "Water", "Dark")
	add(659, "Bunnelby", "掘掘兔", "Normal", "")
	add(660, "Diggersby", "掘地兔", "Normal", "Ground")
	add(661, "Fletchling", "小箭雀", "Normal", "Flying")
	add(662, "Fletchinder", "火箭雀", "Fire", "Flying")
	add(663, "Talonflame", "烈箭鹰", "Fire", "Flying")
	add(664, "Scatterbug", "粉蝶虫", "Bug", "")
	add(665, "Spewpa", "粉蝶蛹", "Bug", "")
	add(666, "Vivillon", "彩粉蝶", "Bug", "Flying")
	add(667, "Litleo", "小狮狮", "Fire", "Normal")
	add(668, "Pyroar", "火炎狮", "Fire", "Normal")
	add(669, "Flabébé", "花蓓蓓", "Fairy", "")
	add(670, "Floette", "花叶蒂", "Fairy", "")
	add(671, "Florges", "花洁夫人", "Fairy", "")
	add(672, "Skiddo", "坐骑小羊", "Grass", "")
	add(673, "Gogoat", "坐骑山羊", "Grass", "")
	add(674, "Pancham", "顽皮熊猫", "Fighting", "")
	add(675, "Pangoro", "流氓熊猫", "Fighting", "Dark")
	add(676, "Furfrou", "多丽米亚", "Normal", "")
	add(677, "Espurr", "妙喵", "Psychic", "")
	add(678, "Meowstic", "超能妙喵", "Psychic", "")
	add(679, "Honedge", "独剑鞘", "Steel", "Ghost")
	add(680, "Doublade", "双剑鞘", "Steel", "Ghost")
	add(681, "Aegislash", "坚盾剑怪", "Steel", "Ghost")
	add(682, "Spritzee", "粉香香", "Fairy", "")
	add(683, "Aromatisse", "芳香精", "Fairy", "")
	add(684, "Swirlix", "绵绵泡芙", "Fairy", "")
	add(685, "Slurpuff", "胖甜妮", "Fairy", "")
	add(686, "Inkay", "好啦鱿", "Dark", "Psychic")
	add(687, "Malamar", "乌贼王", "Dark", "Psychic")
	add(688, "Binacle", "龟脚脚", "Rock", "Water")
	add(689, "Barbaracle", "龟足巨铠", "Rock", "Water")
	add(690, "Skrelp", "垃垃藻", "Poison", "Water")
	add(691, "Dragalge", "毒藻龙", "Poison", "Dragon")
	add(692, "Clauncher", "铁臂枪虾", "Water", "")
	add(693, "Clawitzer", "钢炮臂虾", "Water", "")
	add(694, "Helioptile", "伞电蜥", "Electric", "Normal")
	add(695, "Heliolisk", "光电伞蜥", "Electric", "Normal")
	add(696, "Tyrunt", "宝宝暴龙", "Rock", "Dragon")
	add(697, "Tyrantrum", "怪颚龙", "Rock", "Dragon")
	add(698, "Amaura", "冰雪龙", "Rock", "Ice")
	add(699, "Aurorus", "冰雪巨龙", "Rock", "Ice")
	add(700, "Sylveon", "仙子伊布", "Fairy", "")
	add(701, "Hawlucha", "摔角鹰人", "Fighting", "Flying")
	add(702, "Dedenne", "咚咚鼠", "Electric", "Fairy")
	add(703, "Carbink", "小碎钻", "Rock", "Fairy")
	add(704, "Goomy", "黏黏宝", "Dragon", "")
	add(705, "Sliggoo", "黏美儿", "Dragon", "")
	add(706, "Goodra", "黏美龙", "Dragon", "")
	add(707, "Klefki", "钥圈儿", "Steel", "Fairy")
	add(708, "Phantump", "小木灵", "Ghost", "Grass")
	add(709, "Trevenant", "朽木妖", "Ghost", "Grass")
	add(710, "Pumpkaboo", "南瓜精", "Ghost", "Grass")
	add(711, "Gourgeist", "南瓜怪人", "Ghost", "Grass")
	add(712, "Bergmite", "冰宝", "Ice", "")
	add(713, "Avalugg", "冰岩怪", "Ice", "")
	add(714, "Noibat", "嗡蝠", "Flying", "Dragon")
	add(715, "Noivern", "音波龙", "Flying", "Dragon")
	add(716, "Xerneas", "哲尔尼亚斯", "Fairy", "")
	add(717, "Yveltal", "伊裴尔塔尔", "Dark", "Flying")
	add(718, "Zygarde", "基格尔德", "Dragon", "Ground")
	add(719, "Diancie", "蒂安希", "Rock", "Fairy")
	add(720, "Hoopa", "胡帕", "Psychic", "Ghost")
	add(721, "Volcanion", "波尔凯尼恩", "Fire", "Water")

	// ---------- Generation 7: Alola #722-809 ----------
	add(722, "Rowlet", "木木枭", "Grass", "Flying")
	add(723, "Dartrix", "投羽枭", "Grass", "Flying")
	add(724, "Decidueye", "狙射树枭", "Grass", "Ghost")
	add(725, "Litten", "火斑喵", "Fire", "")
	add(726, "Torracat", "炎热喵", "Fire", "")
	add(727, "Incineroar", "炽焰咆哮虎", "Fire", "Dark")
	add(728, "Popplio", "球球海狮", "Water", "")
	add(729, "Brionne", "花漾海狮", "Water", "")
	add(730, "Primarina", "西狮海壬", "Water", "Fairy")
	add(731, "Pikipek", "小笃儿", "Normal", "Flying")
	add(732, "Trumbeak", "喇叭啄鸟", "Normal", "Flying")
	add(733, "Toucannon", "铳嘴大鸟", "Normal", "Flying")
	add(734, "Yungoos", "猫鼬少", "Normal", "")
	add(735, "Gumshoos", "猫鼬探长", "Normal", "")
	add(736, "Grubbin", "强颚鸡母虫", "Bug", "")
	add(737, "Charjabug", "虫电宝", "Bug", "Electric")
	add(738, "Vikavolt", "锹农炮虫", "Bug", "Electric")
	add(739, "Crabrawler", "好胜蟹", "Fighting", "")
	add(740, "Crabominable", "好胜毛蟹", "Fighting", "Ice")
	add(741, "Oricorio", "花舞鸟", "Fire", "Flying")
	add(742, "Cutiefly", "萌虻", "Bug", "Fairy")
	add(743, "Ribombee", "蝶结萌虻", "Bug", "Fairy")
	add(744, "Rockruff", "岩狗狗", "Rock", "")
	add(745, "Lycanroc", "鬃岩狼人", "Rock", "")
	add(746, "Wishiwashi", "弱丁鱼", "Water", "")
	add(747, "Mareanie", "好坏星", "Poison", "Water")
	add(748, "Toxapex", "超坏星", "Poison", "Water")
	add(749, "Mudbray", "泥驴仔", "Ground", "")
	add(750, "Mudsdale", "重泥挽马", "Ground", "")
	add(751, "Dewpider", "滴蛛", "Water", "Bug")
	add(752, "Araquanid", "滴蛛霸", "Water", "Bug")
	add(753, "Fomantis", "伪螳草", "Grass", "")
	add(754, "Lurantis", "兰螳花", "Grass", "")
	add(755, "Morelull", "睡睡菇", "Grass", "Fairy")
	add(756, "Shiinotic", "灯罩夜菇", "Grass", "Fairy")
	add(757, "Salandit", "夜盗火蜥", "Poison", "Fire")
	add(758, "Salazzle", "焰后蜥", "Poison", "Fire")
	add(759, "Stufful", "童偶熊", "Normal", "Fighting")
	add(760, "Bewear", "穿着熊", "Normal", "Fighting")
	add(761, "Bounsweet", "甜竹竹", "Grass", "")
	add(762, "Steenee", "甜舞妮", "Grass", "")
	add(763, "Tsareena", "甜冷美后", "Grass", "")
	add(764, "Comfey", "花疗环环", "Fairy", "")
	add(765, "Oranguru", "智挥猩", "Normal", "Psychic")
	add(766, "Passimian", "投掷猴", "Fighting", "")
	add(767, "Wimpod", "胆小虫", "Bug", "Water")
	add(768, "Golisopod", "具甲武者", "Bug", "Water")
	add(769, "Sandygast", "沙丘娃", "Ghost", "Ground")
	add(770, "Palossand", "噬沙堡爷", "Ghost", "Ground")
	add(771, "Pyukumuku", "拳海参", "Water", "")
	add(772, "Type: Null", "属性：空", "Normal", "")
	add(773, "Silvally", "银伴战兽", "Normal", "")
	add(774, "Minior", "小陨星", "Rock", "Flying")
	add(775, "Komala", "树枕尾熊", "Normal", "")
	add(776, "Turtonator", "爆焰龟兽", "Fire", "Dragon")
	add(777, "Togedemaru", "托戈德玛尔", "Electric", "Steel")
	add(778, "Mimikyu", "谜拟Ｑ", "Ghost", "Fairy")
	add(779, "Bruxish", "磨牙彩皮鱼", "Water", "Psychic")
	add(780, "Drampa", "老翁龙", "Normal", "Dragon")
	add(781, "Dhelmise", "破破舵轮", "Ghost", "Grass")
	add(782, "Jangmo-o", "心鳞宝", "Dragon", "")
	add(783, "Hakamo-o", "鳞甲龙", "Dragon", "Fighting")
	add(784, "Kommo-o", "杖尾鳞甲龙", "Dragon", "Fighting")
	add(785, "Tapu Koko", "卡璞・鸣鸣", "Electric", "Fairy")
	add(786, "Tapu Lele", "卡璞・蝶蝶", "Psychic", "Fairy")
	add(787, "Tapu Bulu", "卡璞・哞哞", "Grass", "Fairy")
	add(788, "Tapu Fini", "卡璞・鳍鳍", "Water", "Fairy")
	add(789, "Cosmog", "科斯莫古", "Psychic", "")
	add(790, "Cosmoem", "科斯莫姆", "Psychic", "")
	add(791, "Solgaleo", "索尔迦雷欧", "Psychic", "Steel")
	add(792, "Lunala", "露奈雅拉", "Psychic", "Ghost")
	add(793, "Nihilego", "虚吾伊德", "Rock", "Poison")
	add(794, "Buzzwole", "爆肌蚊", "Bug", "Fighting")
	add(795, "Pheromosa", "费洛美螂", "Bug", "Fighting")
	add(796, "Xurkitree", "电束木", "Electric", "")
	add(797, "Celesteela", "铁火辉夜", "Steel", "Flying")
	add(798, "Kartana", "纸御剑", "Grass", "Steel")
	add(799, "Guzzlord", "恶食大王", "Dark", "Dragon")
	add(800, "Necrozma", "奈克洛兹玛", "Psychic", "")
	add(801, "Magearna", "玛机雅娜", "Steel", "Fairy")
	add(802, "Marshadow", "玛夏多", "Fighting", "Ghost")
	add(803, "Poipole", "毒贝比", "Poison", "")
	add(804, "Naganadel", "四颚针龙", "Poison", "Dragon")
	add(805, "Stakataka", "垒磊石", "Rock", "Steel")
	add(806, "Blacephalon", "砰头小丑", "Fire", "Ghost")
	add(807, "Zeraora", "捷拉奥拉", "Electric", "")
	add(808, "Meltan", "美录坦", "Steel", "")
	add(809, "Melmetal", "美录梅塔", "Steel", "")

	// ---------- Generation 8: Galar #810-905 ----------
	add(810, "Grookey", "敲音猴", "Grass", "")
	add(811, "Thwackey", "啪咚猴", "Grass", "")
	add(812, "Rillaboom", "轰擂金刚猩", "Grass", "")
	add(813, "Scorbunny", "炎兔儿", "Fire", "")
	add(814, "Raboot", "腾蹴小将", "Fire", "")
	add(815, "Cinderace", "闪焰王牌", "Fire", "")
	add(816, "Sobble", "泪眼蜥", "Water", "")
	add(817, "Drizzile", "变涩蜥", "Water", "")
	add(818, "Inteleon", "千面避役", "Water", "")
	add(819, "Skwovet", "贪心栗鼠", "Normal", "")
	add(820, "Greedent", "藏饱栗鼠", "Normal", "")
	add(821, "Rookidee", "稚山雀", "Flying", "")
	add(822, "Corvisquire", "蓝鸦", "Flying", "")
	add(823, "Corviknight", "钢铠鸦", "Flying", "Steel")
	add(824, "Blipbug", "索侦虫", "Bug", "")
	add(825, "Dottler", "天罩虫", "Bug", "Psychic")
	add(826, "Orbeetle", "以欧路普", "Bug", "Psychic")
	add(827, "Nickit", "偷儿狐", "Dark", "")
	add(828, "Thievul", "狐大盗", "Dark", "")
	add(829, "Gossifleur", "幼棉棉", "Grass", "")
	add(830, "Eldegoss", "白蓬蓬", "Grass", "")
	add(831, "Wooloo", "毛辫羊", "Normal", "")
	add(832, "Dubwool", "毛毛角羊", "Normal", "")
	add(833, "Chewtle", "咬咬龟", "Water", "")
	add(834, "Drednaw", "暴噬龟", "Water", "Rock")
	add(835, "Yamper", "来电汪", "Electric", "")
	add(836, "Boltund", "逐电犬", "Electric", "")
	add(837, "Rolycoly", "小炭仔", "Rock", "")
	add(838, "Carkol", "大炭车", "Rock", "Fire")
	add(839, "Coalossal", "巨炭山", "Rock", "Fire")
	add(840, "Applin", "啃果虫", "Grass", "Dragon")
	add(841, "Flapple", "苹裹龙", "Grass", "Dragon")
	add(842, "Appletun", "丰蜜龙", "Grass", "Dragon")
	add(843, "Silicobra", "沙包蛇", "Ground", "")
	add(844, "Sandaconda", "沙螺蟒", "Ground", "")
	add(845, "Cramorant", "古月鸟", "Flying", "Water")
	add(846, "Arrokuda", "刺梭鱼", "Water", "")
	add(847, "Barraskewda", "戽斗尖梭", "Water", "")
	add(848, "Toxel", "毒电婴", "Electric", "Poison")
	add(849, "Toxtricity", "毒电兔", "Electric", "Poison")
	add(850, "Sizzlipede", "烧火蚣", "Fire", "Bug")
	add(851, "Centiskorch", "焚焰蚣", "Fire", "Bug")
	add(852, "Clobbopus", "拳拳蛸", "Fighting", "")
	add(853, "Grapploct", "八爪武师", "Fighting", "")
	add(854, "Sinistea", "来悲茶", "Ghost", "")
	add(855, "Polteageist", "怖思壶", "Ghost", "")
	add(856, "Hatenna", "迷布莉姆", "Psychic", "")
	add(857, "Hattrem", "提布莉姆", "Psychic", "")
	add(858, "Hatterene", "布莉姆温", "Psychic", "Fairy")
	add(859, "Impidimp", "捣蛋小妖", "Dark", "Fairy")
	add(860, "Morgrem", "诈唬魔", "Dark", "Fairy")
	add(861, "Grimmsnarl", "长毛巨魔", "Dark", "Fairy")
	add(862, "Obstagoon", "堵拦熊", "Dark", "Normal")
	add(863, "Perrserker", "喵头目", "Steel", "")
	add(864, "Cursola", "魔灵珊瑚", "Ghost", "")
	add(865, "Sirfetch'd", "葱游兵", "Fighting", "")
	add(866, "Mr. Rime", "踏冰人偶", "Ice", "Psychic")
	add(867, "Runerigus", "死神板", "Ground", "Ghost")
	add(868, "Milcery", "小仙奶", "Fairy", "")
	add(869, "Alcremie", "霜奶仙", "Fairy", "")
	add(870, "Falinks", "列阵兵", "Fighting", "")
	add(871, "Pincurchin", "啪嚓海胆", "Electric", "")
	add(872, "Snom", "雪吞虫", "Ice", "Bug")
	add(873, "Frosmoth", "雪绒蛾", "Ice", "Bug")
	add(874, "Stonjourner", "巨石丁", "Rock", "")
	add(875, "Eiscue", "冰砌鹅", "Ice", "")
	add(876, "Indeedee", "爱管侍", "Psychic", "Normal")
	add(877, "Morpeko", "莫鲁贝可", "Electric", "Dark")
	add(878, "Cufant", "铜象", "Steel", "")
	add(879, "Copperajah", "大王铜象", "Steel", "")
	add(880, "Dracozolt", "雷鸟龙", "Electric", "Dragon")
	add(881, "Arctozolt", "雷鸟海兽", "Electric", "Ice")
	add(882, "Dracovish", "鳃鱼龙", "Water", "Dragon")
	add(883, "Arctovish", "鳃鱼海兽", "Water", "Ice")
	add(884, "Duraludon", "铝钢龙", "Steel", "Dragon")
	add(885, "Dreepy", "多龙梅西亚", "Dragon", "Ghost")
	add(886, "Drakloak", "多龙奇", "Dragon", "Ghost")
	add(887, "Dragapult", "多龙巴鲁托", "Dragon", "Ghost")
	add(888, "Zacian", "苍响", "Fairy", "")
	add(889, "Zamazenta", "藏玛然特", "Fighting", "")
	add(890, "Eternatus", "无极汰那", "Poison", "Dragon")
	add(891, "Kubfu", "熊徒弟", "Fighting", "")
	add(892, "Urshifu", "武道熊师", "Fighting", "Dark")
	add(893, "Zarude", "萨戮德", "Dark", "Grass")
	add(894, "Regieleki", "雷吉艾勒奇", "Electric", "")
	add(895, "Regidrago", "雷吉铎拉戈", "Dragon", "")
	add(896, "Glastrier", "雪暴马", "Ice", "")
	add(897, "Spectrier", "幽灵马", "Ghost", "")
	add(898, "Calyrex", "蕾冠王", "Psychic", "Grass")

	// ---------- Generation 8: Hisui #899-905 ----------
	add(899, "Wyrdeer", "诡角鹿", "Normal", "Psychic")
	add(900, "Kleavor", "劈斧螳螂", "Bug", "Rock")
	add(901, "Ursaluna", "月月熊", "Ground", "Normal")
	add(902, "Basculegion", "幽尾玄鱼", "Water", "Ghost")
	add(903, "Sneasler", "大狃拉", "Fighting", "Poison")
	add(904, "Overqwil", "万针鱼", "Dark", "Poison")
	add(905, "Enamorus", "眷恋云", "Fairy", "Flying")

	// ---------- Generation 9: Paldea #906-1025 ----------
	add(906, "Sprigatito", "新叶喵", "Grass", "")
	add(907, "Floragato", "蒂蕾喵", "Grass", "")
	add(908, "Meowscarada", "魔幻假面喵", "Grass", "Dark")
	add(909, "Fuecoco", "呆火鳄", "Fire", "")
	add(910, "Crocalor", "炙烫鳄", "Fire", "")
	add(911, "Skeledirge", "骨纹巨声鳄", "Fire", "Ghost")
	add(912, "Quaxly", "润水鸭", "Water", "")
	add(913, "Quaxwell", "涌跃鸭", "Water", "")
	add(914, "Quaquaval", "狂欢浪舞鸭", "Water", "Fighting")
	add(915, "Lechonk", "爱吃豚", "Normal", "")
	add(916, "Oinkologne", "飘香豚", "Normal", "")
	add(917, "Tarountula", "团珠蛛", "Bug", "")
	add(918, "Spidops", "操陷蛛", "Bug", "")
	add(919, "Nymble", "豆蟋蟀", "Bug", "")
	add(920, "Lokix", "烈腿蝗", "Bug", "Dark")
	add(921, "Pawmi", "布拨", "Electric", "")
	add(922, "Pawmo", "布土拨", "Electric", "Fighting")
	add(923, "Pawmot", "巴布土拨", "Electric", "Fighting")
	add(924, "Tandemaus", "一对鼠", "Normal", "")
	add(925, "Maushold", "一家鼠", "Normal", "")
	add(926, "Fidough", "狗仔包", "Fairy", "")
	add(927, "Dachsbun", "麻花犬", "Fairy", "")
	add(928, "Smoliv", "迷你芙", "Grass", "Normal")
	add(929, "Dolliv", "奥利纽", "Grass", "Normal")
	add(930, "Arboliva", "奥利瓦", "Grass", "Normal")
	add(931, "Squawkabilly", "怒鹦哥", "Normal", "Flying")
	add(932, "Nacli", "盐石宝", "Rock", "")
	add(933, "Naclstack", "盐石垒", "Rock", "")
	add(934, "Garganacl", "盐石巨灵", "Rock", "")
	add(935, "Charcadet", "炭小侍", "Fire", "")
	add(936, "Armarouge", "红莲铠骑", "Fire", "Psychic")
	add(937, "Ceruledge", "苍炎刃鬼", "Fire", "Ghost")
	add(938, "Tadbulb", "光蚪仔", "Electric", "")
	add(939, "Bellibolt", "电肚蛙", "Electric", "")
	add(940, "Wattrel", "电海燕", "Electric", "Flying")
	add(941, "Kilowattrel", "大电海燕", "Electric", "Flying")
	add(942, "Maschiff", "獒教父", "Dark", "")
	add(943, "Mabosstiff", "獒教父", "Dark", "")
	add(944, "Shroodle", "滋汁鼹", "Poison", "Normal")
	add(945, "Grafaiai", "涂标客", "Poison", "Normal")
	add(946, "Bramblin", "纳噬草", "Grass", "Ghost")
	add(947, "Brambleghast", "怖纳噬草", "Grass", "Ghost")
	add(948, "Toedscool", "原野水母", "Ground", "Grass")
	add(949, "Toedscruel", "陆地水母", "Ground", "Grass")
	add(950, "Klawf", "毛崖蟹", "Rock", "")
	add(951, "Capsakid", "热辣娃", "Grass", "")
	add(952, "Scovillain", "狠辣椒", "Grass", "Fire")
	add(953, "Rellor", "虫滚泥", "Bug", "")
	add(954, "Rabsca", "虫甲圣", "Bug", "Psychic")
	add(955, "Flittle", "飘飘雏", "Psychic", "")
	add(956, "Espathra", "超能艳鸵", "Psychic", "")
	add(957, "Tinkatink", "小锻匠", "Fairy", "Steel")
	add(958, "Tinkatuff", "巧锻匠", "Fairy", "Steel")
	add(959, "Tinkaton", "巨锻匠", "Fairy", "Steel")
	add(960, "Wiglett", "海地鼠", "Water", "")
	add(961, "Wugtrio", "三海地鼠", "Water", "")
	add(962, "Bombirdier", "下石鸟", "Flying", "Dark")
	add(963, "Finizen", "波普海豚", "Water", "")
	add(964, "Palafin", "海豚侠", "Water", "")
	add(965, "Varoom", "噗隆隆", "Steel", "Poison")
	add(966, "Revavroom", "普隆隆姆", "Steel", "Poison")
	add(967, "Cyclizar", "摩托蜥", "Dragon", "Normal")
	add(968, "Orthworm", "拖拖蚓", "Steel", "")
	add(969, "Glimmet", "晶光芽", "Rock", "Poison")
	add(970, "Glimmora", "晶光花", "Rock", "Poison")
	add(971, "Greavard", "墓仔狗", "Ghost", "")
	add(972, "Houndstone", "墓扬犬", "Ghost", "")
	add(973, "Flamigo", "缠红鹤", "Flying", "Fighting")
	add(974, "Cetoddle", "走鲸", "Ice", "")
	add(975, "Cetitan", "浩大鲸", "Ice", "")
	add(976, "Veluza", "轻身鳕", "Water", "Psychic")
	add(977, "Dondozo", "吃吼霸", "Water", "")
	add(978, "Tatsugiri", "米立龙", "Dragon", "Water")
	add(979, "Annihilape", "弃世猴", "Fighting", "Ghost")
	add(980, "Clodsire", "土王", "Poison", "Ground")
	add(981, "Farigiraf", "奇麒麟", "Normal", "Psychic")
	add(982, "Dudunsparce", "土龙节节", "Normal", "")
	add(983, "Kingambit", "仆刀将军", "Dark", "Steel")
	add(984, "Great Tusk", "雄伟牙", "Ground", "Fighting")
	add(985, "Scream Tail", "吼叫尾", "Fairy", "Psychic")
	add(986, "Brute Bonnet", "猛恶菇", "Grass", "Dark")
	add(987, "Flutter Mane", "振翼发", "Ghost", "Fairy")
	add(988, "Slither Wing", "爬地翅", "Bug", "Fighting")
	add(989, "Sandy Shocks", "沙铁皮", "Electric", "Ground")
	add(990, "Iron Treads", "铁辙迹", "Ground", "Steel")
	add(991, "Iron Bundle", "铁包袱", "Ice", "Water")
	add(992, "Iron Hands", "铁臂膀", "Fighting", "Electric")
	add(993, "Iron Jugulis", "铁脖颈", "Dark", "Flying")
	add(994, "Iron Moth", "铁毒蛾", "Fire", "Poison")
	add(995, "Iron Thorns", "铁荆棘", "Rock", "Electric")
	add(996, "Frigibax", "凉脊龙", "Dragon", "Ice")
	add(997, "Arctibax", "冻脊龙", "Dragon", "Ice")
	add(998, "Baxcalibur", "戟脊龙", "Dragon", "Ice")
	add(999, "Gimmighoul", "索财灵", "Ghost", "")
	add(1000, "Gholdengo", "赛富豪", "Steel", "Ghost")
	add(1001, "Wo-Chien", "古简蜗", "Dark", "Grass")
	add(1002, "Chien-Pao", "古剑豹", "Dark", "Ice")
	add(1003, "Ting-Lu", "古鼎鹿", "Dark", "Ground")
	add(1004, "Chi-Yu", "古玉鱼", "Dark", "Fire")
	add(1005, "Roaring Moon", "轰鸣月", "Dragon", "Dark")
	add(1006, "Iron Valiant", "铁武者", "Fairy", "Fighting")
	add(1007, "Koraidon", "故勒顿", "Fighting", "Dragon")
	add(1008, "Miraidon", "密勒顿", "Electric", "Dragon")
	add(1009, "Walking Wake", "波荡水", "Water", "Dragon")
	add(1010, "Iron Leaves", "铁斑叶", "Grass", "Psychic")
	add(1011, "Dipplin", "裹蜜虫", "Grass", "Dragon")
	add(1012, "Poltchageist", "斯魔茶", "Grass", "Ghost")
	add(1013, "Sinistcha", "来悲粗茶", "Grass", "Ghost")
	add(1014, "Okidogi", "够赞狗", "Poison", "Fighting")
	add(1015, "Munkidori", "愿增猿", "Poison", "Psychic")
	add(1016, "Fezandipiti", "吉雉鸡", "Poison", "Fairy")
	add(1017, "Ogerpon", "厄诡椪", "Grass", "")
	add(1018, "Archaludon", "铝钢桥龙", "Steel", "Dragon")
	add(1019, "Hydrapple", "蜜集大蛇", "Grass", "Dragon")
	add(1020, "Gouging Fire", "破空火", "Fire", "Dragon")
	add(1021, "Raging Bolt", "猛雷鼓", "Electric", "Dragon")
	add(1022, "Iron Boulder", "铁磐岩", "Rock", "Psychic")
	add(1023, "Iron Crown", "铁头壳", "Steel", "Psychic")
	add(1024, "Terapagos", "太乐巴戈斯", "Normal", "")
	add(1025, "Pecharunt", "桃歹郎", "Poison", "Ghost")

	return pokemon
}

