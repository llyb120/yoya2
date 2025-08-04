package test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/llyb120/yoya2/y"
)

func TestCollect(t *testing.T) {
	// src := "name[age=10,name*=张三] id [1,2,3] name[age=10,name='张三']"

	data := map[string]any{
		"name": []any{
			map[string]any{
				"age":  10,
				"name": "张三",
				"id":   1,
				"child": []map[string]any{
					{
						"child": 1,
					},
				},
			},
			map[string]any{
				"age":  10,
				"name": "张三",
				"id":   1,
			},
		},
	}
	results := y.Pick[string](data, "name [age=10,name='张三'] id", y.UseDistinct)
	fmt.Printf("%+v\n", results)

	reuslt2 := y.Pick[any](data, "child")
	fmt.Printf("%+v\n", reuslt2)
}

func TestComplexPick(t *testing.T) {
	// 创建一个复杂的嵌套数据结构
	complexData := map[string]any{
		"users": []map[string]any{
			{
				"id": 1,
				"profile": map[string]any{
					"name": "张三",
					"age":  28,
					"skills": []map[string]any{
						{"name": "编程", "level": 9},
						{"name": "设计", "level": 7},
					},
					"contact": map[string]any{
						"email": "zhangsan@example.com",
						"phone": "13800138000",
					},
				},
				"posts": []map[string]any{
					{
						"id":      101,
						"title":   "如何学习Go语言",
						"content": "Go语言是一门很棒的语言...",
						"tags":    []string{"Go", "编程", "学习"},
						"comments": []map[string]any{
							{"user": "李四", "content": "非常有用的文章"},
							{"user": "王五", "content": "谢谢分享"},
						},
					},
					{
						"id":      102,
						"title":   "数据结构基础",
						"content": "理解数据结构对编程很重要...",
						"tags":    []string{"数据结构", "编程", "基础"},
						"comments": []map[string]any{
							{"user": "赵六", "content": "讲解得很清楚"},
						},
					},
				},
			},
			{
				"id": 2,
				"profile": map[string]any{
					"name": "李四",
					"age":  32,
					"skills": []map[string]any{
						{"name": "管理", "level": 8},
						{"name": "编程", "level": 6},
					},
					"contact": map[string]any{
						"email": "lisi@example.com",
						"phone": "13900139000",
					},
				},
				"posts": []map[string]any{
					{
						"id":      201,
						"title":   "项目管理技巧",
						"content": "有效的项目管理需要...",
						"tags":    []string{"管理", "项目", "技巧"},
						"comments": []map[string]any{
							{"user": "张三", "content": "学到了很多"},
						},
					},
				},
			},
		},
		"categories": []map[string]any{
			{"id": 1, "name": "技术"},
			{"id": 2, "name": "管理"},
			{"id": 3, "name": "设计"},
		},
	}

	// 测试1：查找所有技能等级大于7的技能
	skills := y.Pick[any](complexData, "[level>7] level")
	fmt.Println("高级技能:")
	for _, skill := range skills {
		fmt.Println(skill)
		// fmt.Printf("  %s (等级: %v)\n", skill["name"], skill["level"])
	}

	// 测试2：查找所有张三的文章评论
	comments := y.Pick[map[string]any](complexData, "comments [user='张三']")
	fmt.Println("\n张三的评论:")
	for _, comment := range comments {
		fmt.Printf("  %s: %s\n", comment["user"], comment["content"])
	}

	// 测试3：查找所有带有"编程"标签的文章
	posts := y.Pick[map[string]any](complexData, "posts")
	fmt.Println("\n编程相关文章:")
	for _, post := range posts {
		tags, ok := post["tags"].([]string)
		if ok {
			for _, tag := range tags {
				if tag == "编程" {
					fmt.Printf("  %s\n", post["title"])
					break
				}
			}
		}
	}

	// 测试4：查找所有用户的联系方式
	contacts := y.Pick[map[string]any](complexData, "contact")
	fmt.Println("\n用户联系方式:")
	for i, contact := range contacts {
		fmt.Printf("  用户%d: 邮箱=%s, 电话=%s\n", i+1, contact["email"], contact["phone"])
	}

	// 测试5：使用多层嵌套选择器
	userPosts := y.Pick[map[string]any](complexData, "users profile[name='张三'] posts")
	fmt.Println("\n张三的所有文章:")
	for _, post := range userPosts {
		fmt.Printf("  %s\n", post["title"])
	}

	// 测试6：复杂条件组合
	result := y.Pick[map[string]any](complexData, "users [id=1] profile skills [level>5]")
	fmt.Println("\nID为1的用户的高级技能(等级>5):")
	for _, item := range result {
		fmt.Printf("  %s (等级: %v)\n", item["name"], item["level"])
	}

	// 测试7：使用Walk进行数据转换
	// fmt.Println("\n将所有年龄增加1:")
	// Walk(complexData, func(s any, k any, v any) any {
	// 	if k == "age" {
	// 		if age, ok := v.(int); ok {
	// 			return age + 1
	// 		}
	// 	}
	// 	return Unchanged
	// })

	// 验证年龄是否已更新
	users := y.Pick[map[string]any](complexData, "profile")
	for _, user := range users {
		fmt.Printf("  %s: %d岁\n", user["name"], user["age"])
	}
}

func TestPickSort(t *testing.T) {
	// LANGUAGE=json
	jsonStr := `
	{
    "code": 0,
    "msg": "ok",
    "system": "intelligence",
    "data": {
        "version": "2.0",
        "xAxis": null,
        "yAxis": null,
        "legends": null,
        "chat_type": "table",
        "metrics_info": [
            {
                "name": "收入(美元)",
                "data_key": "revenue",
                "type": "numerical"
            },
            {
                "name": "",
                "data_key": "revenue_radio",
                "type": "percent"
            },
            {
                "name": "下载",
                "data_key": "download",
                "type": "numerical"
            },
            {
                "name": "",
                "data_key": "download_radio",
                "type": "percent"
            }
        ],
        "dimension_info": [
            {
                "name": "",
                "data_key": "source",
                "value": {
                    "sensortower": {
                        "name": "Sensor Tower",
                        "value": "sensortower"
                    }
                }
            },
            {
                "name": "",
                "data_key": "game_type",
                "value": [
                    "mobile"
                ]
            },
            {
                "name": "",
                "data_key": "granularity",
                "value": [
                    "monthly"
                ]
            },
            {
                "name": "",
                "data_key": "market_name",
                "value": [
                    "德国",
                    "巴西",
                    "俄罗斯",
                    "越南",
                    "日本",
                    "印度",
                    "巴基斯坦",
                    "埃及",
                    "泰国",
                    "全球（除中国大陆）",
                    "英国",
                    "中国台湾",
                    "意大利",
                    "美国",
                    "土耳其",
                    "沙特阿拉伯",
                    "印度尼西亚",
                    "伊拉克",
                    "韩国",
                    "荷兰",
                    "瑞士",
                    "西班牙",
                    "墨西哥",
                    "澳大利亚",
                    "法国",
                    "全球",
                    "孟加拉",
                    "中国",
                    "菲律宾",
                    "阿根廷",
                    "哥伦比亚",
                    "中国香港",
                    "加拿大"
                ]
            }
        ],
        "filter_info": null,
        "ext_info": {
            "reference_urls": [
                {
                    "title": "Mobile Market Profile",
                    "url": "v2/intelligence/marketProfile/MobileGames"
                }
            ]
        },
        "data": [
            {
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 11238893566,
                "revenue_radio": -0.0405,
                "market_name": "全球（除中国大陆）",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 13209197213,
                "revenue_radio": -0.0823,
                "market_name": "全球",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 1,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "market_name": "美国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "revenue": 4333586634,
                "revenue_radio": -0.021
            },
            {
                "rank": 2,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 1970303647,
                "revenue_radio": -0.2648,
                "market_name": "中国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 3,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 1756908582,
                "revenue_radio": -0.1943,
                "market_name": "日本",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 4,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue_radio": 0.0834,
                "market_name": "韩国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "revenue": 815159284
            },
            {
                "rank": 5,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 449213439,
                "revenue_radio": -0.0328,
                "market_name": "德国",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 6,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 419454230,
                "revenue_radio": 0.0816,
                "market_name": "英国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 7,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 415915868,
                "revenue_radio": -0.0891,
                "market_name": "中国台湾",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 8,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "source": "sensortower",
                "granularity": "monthly",
                "revenue": 277417730,
                "revenue_radio": -0.008,
                "market_name": "加拿大",
                "game_type": "mobile"
            },
            {
                "rank": 9,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 240847267,
                "revenue_radio": -0.0421,
                "market_name": "法国",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 10,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 232158515,
                "revenue_radio": 0.0342,
                "market_name": "澳大利亚",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 11,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "source": "sensortower",
                "granularity": "monthly",
                "revenue": 142200881,
                "revenue_radio": -0.062,
                "market_name": "中国香港",
                "game_type": "mobile"
            },
            {
                "rank": 12,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "market_name": "意大利",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "revenue": 128663569,
                "revenue_radio": 0.017
            },
            {
                "rank": 13,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 115750021,
                "revenue_radio": 0.0235,
                "market_name": "泰国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 14,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 113460642,
                "revenue_radio": 0.3778,
                "market_name": "墨西哥",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 15,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 97363677,
                "revenue_radio": -0.0161,
                "market_name": "巴西",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 16,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "revenue": 93768876,
                "revenue_radio": -0.0774,
                "market_name": "土耳其"
            },
            {
                "rank": 17,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "revenue": 84695072,
                "revenue_radio": 0.0824,
                "market_name": "荷兰",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 18,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 84572308,
                "revenue_radio": -0.1089,
                "market_name": "瑞士",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 19,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 79595079,
                "revenue_radio": 0.0395,
                "market_name": "西班牙",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 20,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "revenue": 78466189,
                "revenue_radio": -0.0375,
                "market_name": "沙特阿拉伯",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 11115724178,
                "download_radio": -0.0941,
                "market_name": "全球（除中国大陆）",
                "game_type": "mobile"
            },
            {
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 11291484205,
                "download_radio": -0.0943,
                "market_name": "全球",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 1,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 2066818205,
                "download_radio": 0.0033,
                "market_name": "印度",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 2,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 845677016,
                "download_radio": -0.1227,
                "market_name": "美国"
            },
            {
                "rank": 3,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 757993047,
                "download_radio": -0.1284,
                "market_name": "印度尼西亚"
            },
            {
                "rank": 4,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 740200372,
                "download_radio": -0.1994,
                "market_name": "巴西",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 5,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 445165189,
                "download_radio": -0.1998,
                "market_name": "墨西哥",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 6,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 434589408,
                "download_radio": -0.2998,
                "market_name": "俄罗斯",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 7,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 348860263,
                "download_radio": -0.005,
                "market_name": "巴基斯坦",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 8,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 336285674,
                "download_radio": -0.346,
                "market_name": "土耳其",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 9,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download_radio": -0.1011,
                "download": 329217122,
                "market_name": "菲律宾",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 10,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 302683180,
                "download_radio": -0.0809,
                "market_name": "越南",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 11,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download_radio": null,
                "market_name": "伊拉克",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 261640529
            },
            {
                "rank": 12,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "download": 253780675,
                "download_radio": null,
                "market_name": "孟加拉",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 13,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 197500971,
                "download_radio": -0.1969,
                "market_name": "埃及",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 14,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "granularity": "monthly",
                "download": 192224911,
                "download_radio": 0.0653,
                "market_name": "泰国",
                "game_type": "mobile",
                "source": "sensortower"
            },
            {
                "rank": 15,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 175760027,
                "download_radio": -0.1077,
                "market_name": "中国",
                "game_type": "mobile"
            },
            {
                "rank": 16,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 158281995,
                "download_radio": -0.1541,
                "market_name": "阿根廷",
                "game_type": "mobile"
            },
            {
                "rank": 17,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "market_name": "德国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 150198781,
                "download_radio": -0.2317
            },
            {
                "rank": 18,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "download": 148009225,
                "download_radio": -0.1972,
                "market_name": "法国",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly"
            },
            {
                "rank": 19,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 144285732,
                "download_radio": -0.1013,
                "market_name": "日本"
            },
            {
                "rank": 20,
                "start_date": "2025-04-01",
                "end_date": "2025-06-01",
                "market_type": "market",
                "market_name": "哥伦比亚",
                "game_type": "mobile",
                "source": "sensortower",
                "granularity": "monthly",
                "download": 141510133,
                "download_radio": -0.2006
            }
        ]
    }
}`

	var obj map[string]any
	json.Unmarshal([]byte(jsonStr), &obj)
	fmt.Printf("%+v\n", obj)
	src := "data market_name"
	results := y.Pick[string](obj, src)
	fmt.Printf("%+v\n", results)
}
