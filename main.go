package main

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	// ==========================================
	// 1. 基础设置（只执行一次）
	// ==========================================
	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	// 获取当前页面对象
	// 如果你手动导航，page 对象通常会自动跟随当前标签页
	page := browser.MustPage("http://teachingevaluationsystem.cn/courses")

	fmt.Println(">>> 浏览器已启动...")
	fmt.Println(">>> 步骤1：请在浏览器中【手动登录】。")
	fmt.Println(">>> 步骤2：请手动进入【第一个】需要评价的老师页面。")
	fmt.Println(">>> 步骤3：准备好后，回到这里按【回车】开始...")
	fmt.Scanln()

	// ==========================================
	// 2. 外层循环：针对“每个老师”
	// ==========================================
	teacherCounter := 1

	for {
		fmt.Printf("\n========== 正在开始第 %d 位老师的自动评价 ==========\n", teacherCounter)

		// --- 每次开始新老师时，重置页码计数器 ---
		pageCounter := 1

		// ==========================================
		// 3. 内层循环：针对“每一页”
		// ==========================================
		for {
			fmt.Printf("   [第 %d 页] 正在扫描选项...\n", pageCounter)

			// --------------------------------------
			// A. 查找并勾选当前页的所有选项
			// --------------------------------------
			xpath := "//span[contains(text(), '非常满意')]"

			// 使用 Race 机制来防止页面加载慢导致的报错，或者页面其实已经没有题目了
			// 这里简单处理：直接找元素
			elements, err := page.ElementsX(xpath)

			if err != nil || len(elements) == 0 {
				fmt.Println("   ! 未找到“非常满意”选项 (可能是网络卡顿或已无题目)。")
			} else {
				fmt.Printf("   -> 发现 %d 个题目，正在执行 JS 点击...\n", len(elements))
				for _, el := range elements {
					// 忽略错误，强制点击
					func() {
						defer func() { recover() }()
						el.MustEval("() => this.click()")
					}()
					time.Sleep(20 * time.Millisecond) //稍微快一点
				}
			}

			// --------------------------------------
			// B. 查找“下一页”并翻页
			// --------------------------------------
			// 使用你刚才测试成功的那个 XPath（如果不确定，就用下面这个宽容模式）
			nextBtnXPath := "//*[contains(., '下一')]"

			hasNext, nextBtn, _ := page.HasX(nextBtnXPath)

			if hasNext {
				fmt.Println("   -> 检测到【下一页/下一步】，正在翻页...")

				// 点击翻页
				nextBtn.MustEval("() => this.click()")

				// 等待新页面加载 (重要！)
				time.Sleep(1 * time.Second)
				pageCounter++
			} else {
				// 没有下一页了，说明这个老师评完了
				fmt.Println("   -> 未检测到下一页按钮，本位老师评价结束。")
				break // 跳出内层循环 (Page Loop)
			}
		}

		// ==========================================
		// 4. 单个老师结束后的交互
		// ==========================================
		fmt.Println("\n✅ 第", teacherCounter, "位老师评价完成！")
		fmt.Println("--------------------------------------------------")
		fmt.Println("请执行以下操作：")
		fmt.Println("1. 在浏览器手动点击【提交】。")
		fmt.Println("2. 点击【返回】列表。")
		fmt.Println("3. 点击进入【下一位老师】的页面。")
		fmt.Println("--------------------------------------------------")
		fmt.Println(">>> 准备好评下一位了吗？按【回车】继续 (输入 q 并回车退出)...")

		var input string
		fmt.Scanln(&input) // 等待用户按回车

		if input == "q" {
			fmt.Println("退出程序。辛苦了！")
			break // 跳出外层循环，结束程序
		}

		teacherCounter++
		// 循环回到开头，pageCounter 会被重置，继续评下一个
	}
}
