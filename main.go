package main

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// 配置区：请根据实际情况修改
const (
	// 评教列表的主页地址
	MainListURL = "http://teachingevaluationsystem.cn/courses"

	// 列表页：进入评价的按钮特征 (比如 "去评价", "未评", "评价")
	// 建议找那些状态为“未完成”的按钮上的字
	EntryBtnText = "评价"

	// 详情页：最后提交按钮的文字
	SubmitBtnText = "提交"

	// 详情页：确认弹窗的按钮文字 (Vant UI 点击提交后通常会有个“确认”弹窗)
	ConfirmBtnText = "确定"
)

func main() {
	// 1. 初始化
	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage(MainListURL)

	fmt.Println(">>> 浏览器已启动...")
	fmt.Println(">>> 请手动登录。登录成功后，请确保页面停留在【评教列表页】。")
	fmt.Println(">>> 确认可以开始后，按【回车】启动全自动模式...")
	fmt.Scanln()

	// 2. 死循环：只要列表页还有“去评价”的按钮，就一直干活
	for {
		fmt.Println("\n========== 正在检查列表页任务 ==========")

		// 确保我们在列表页
		if page.MustInfo().URL != MainListURL {
			page.MustNavigate(MainListURL)
			page.MustWaitLoad()
			time.Sleep(2 * time.Second) // 等列表加载
		}

		// 查找第一个可用的入口按钮
		// 这里使用模糊匹配：查找所有包含 EntryBtnText (例如"评价") 的元素
		// 并且它应该是 clickable 的
		entryXPath := fmt.Sprintf("//*[contains(text(), '%s')]", EntryBtnText)

		// 搜索页面上是否还有这个按钮
		// 我们只拿第一个 (First)，评完一个回来再拿下一个
		hasTask, entryBtn, _ := page.HasX(entryXPath)

		if !hasTask {
			fmt.Println("列表页找不到包含“" + EntryBtnText + "”的按钮了。")
			fmt.Println("恭喜！所有评价可能已完成。")
			break // 结束整个程序
		}

		fmt.Println("发现待评价课程，准备进入...")

		// 点击进入详情页
		entryBtn.MustEval("() => this.click()")

		// 等待详情页加载
		time.Sleep(2 * time.Second)

		// === 执行单个老师的评价逻辑 ===
		processCurrentTeacher(page)

		// 评价完一个后，循环会回到开头，重新刷新列表页，找下一个
	}
}

// processCurrentTeacher 处理单个老师的所有页面和提交
func processCurrentTeacher(page *rod.Page) {
	pageCounter := 1

	for {
		fmt.Printf("   [第 %d 页] 正在自动勾选...\n", pageCounter)

		// 1. 勾选“非常满意” (复用之前的逻辑)
		xpath := "//span[contains(text(), '非常满意')]"
		// 使用 Race 避免找不到元素报错
		if elements, err := page.ElementsX(xpath); err == nil {
			for _, el := range elements {
				// 忽略错误，强制点击
				func() {
					defer func() { recover() }()
					el.MustEval("() => this.click()")
				}()
				time.Sleep(20 * time.Millisecond)
			}
		}

		// 2. 判断是“翻页”还是“提交”
		// 先找下一页
		nextBtnXPath := "//*[contains(., '下一')]" // 你的翻页按钮文字
		hasNext, nextBtn, _ := page.HasX(nextBtnXPath)

		if hasNext {
			// --- 有下一页，继续翻 ---
			fmt.Println("   -> 翻页...")
			nextBtn.MustEval("() => this.click()")
			time.Sleep(1 * time.Second)
			pageCounter++
		} else {
			// --- 没有下一页了，说明到底了，找提交按钮 ---
			fmt.Println("   -> 已到最后一页，查找提交按钮...")

			submitXPath := fmt.Sprintf("//button[contains(., '%s')] | //div[contains(., '%s') and contains(@class, 'btn')]", SubmitBtnText, SubmitBtnText)

			hasSubmit, submitBtn, _ := page.HasX(submitXPath)
			if hasSubmit {
				fmt.Println("   -> 点击提交！")
				submitBtn.MustEval("() => this.click()")

				// 3. 处理可能出现的“确认提交”弹窗
				// Vant UI 通常会弹出一个 Dialog 让你点“确定”
				time.Sleep(1 * time.Second) // 等弹窗出来

				confirmXPath := fmt.Sprintf("//button[contains(., '%s')]", ConfirmBtnText)
				if hasConfirm, confirmBtn, _ := page.HasX(confirmXPath); hasConfirm {
					fmt.Println("   -> 确认弹窗...")
					confirmBtn.MustClick()
				}

				fmt.Println("✅ 本课程评价完成！等待返回列表...")
				time.Sleep(3 * time.Second) // 等待提交请求完成并跳转
				return                      // 退出函数，回到 Main 的列表循环
			} else {
				fmt.Println("❌ 既没有下一页，也没找到提交按钮，可能出错了。")
				// 这种情况下，为了防止死循环，我们强制回退
				return
			}
		}
	}
}
