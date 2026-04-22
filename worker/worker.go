// ============================================================
// worker/worker.go — Goroutine + Database Lock 搶票教學核心
// ============================================================
//
// 這個檔案用「搶票」的情境，示範兩種做法：
//   1. 不加鎖的搶票（會超賣！觀察 race condition）
//   2. 加 SELECT FOR UPDATE 鎖的搶票（正確！不會超賣）
//
// 搭配 /rush/without-lock 和 /rush/with-lock 兩支 API 呼叫，
// 可以直接從回傳的結果看出「有鎖 vs 沒鎖」的差異。
//
// ─────────────────────────────────────────────────────────
// 【概念一：什麼是 Race Condition（競爭條件）？】
//
//   假設活動剩 5 張票，同時有 10 個人搶票：
//   - 沒有鎖的情況：10 個 goroutine「同時」查庫存，大家都看到還有 5 張
//     → 每個人都以為自己搶得到 → 結果賣出超過 5 張 → 超賣！
//
//   這就是 race condition：多個 goroutine 同時讀寫同一筆資料，
//   導致結果不符合預期。
//
// ─────────────────────────────────────────────────────────
// 【概念二：SELECT ... FOR UPDATE 是什麼？】
//
//   在 transaction 裡面用 SELECT ... FOR UPDATE 查詢某一行：
//   - PostgreSQL 會把那一行「鎖住」
//   - 其他 transaction 如果也想 SELECT FOR UPDATE 同一行，就必須等
//   - 等到第一個 transaction COMMIT 或 ROLLBACK 後，鎖才會釋放
//
//   這樣就能確保「查庫存 → 扣庫存」這兩步是不可分割的：
//     BEGIN
//     SELECT stock FROM events WHERE id = 1 FOR UPDATE;  -- 鎖住這行
//     -- 這時候其他人想讀同一行會被擋住，等我做完
//     UPDATE events SET stock = stock - 1 WHERE id = 1;
//     COMMIT;  -- 鎖釋放，下一個人才能進來
//
// ─────────────────────────────────────────────────────────
// 【概念三：sync.WaitGroup 複習】
//
//   WaitGroup 就像「等候區的號碼牌計數器」：
//     - wg.Add(1)  → 計數器 +1（告訴主程式：又多一個人要忙）
//     - wg.Done()  → 計數器 -1（告訴主程式：我忙完了）
//     - wg.Wait()  → 等到計數器歸零（所有人都忙完）才繼續
//
// ─────────────────────────────────────────────────────────
// 【概念四：goroutine 的閉包陷阱】
//
//   for i := 0; i < 5; i++ {
//       go func() {
//           fmt.Println(i)  // ⚠️ 錯！goroutine 真正跑的時候 i 可能已經變成 5 了
//       }()
//   }
//
//   正確做法：把 i 當參數傳進去
//   for i := 0; i < 5; i++ {
//       go func(buyerID int) {
//           fmt.Println(buyerID)  // ✅ 對！每個 goroutine 都有自己的 copy
//       }(i)
//   }
//
// ============================================================

package worker

import (
	"fmt"
	"sync" // TODO: 實作 goroutine 時會用到 sync.WaitGroup
	"time"

	"go-api-practice-10/database"
)

// 避免 import 報錯（實作完成後可移除這行）
var _ = sync.WaitGroup{}

// RushRequest 是搶票模擬的請求參數
type RushRequest struct {
	EventID int `json:"event_id" binding:"required,gte=1"`        // 哪個活動
	Buyers  int `json:"buyers"   binding:"required,gte=1,lte=50"` // 幾個人同時搶
}

// BuyerResult 是單一買家的搶票結果
type BuyerResult struct {
	BuyerID int    `json:"buyer_id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// RushResult 是整場搶票的總結果
type RushResult struct {
	Type        string        `json:"type"`          // 模擬類型
	EventID     int           `json:"event_id"`      // 活動 ID
	TotalBuyers int           `json:"total_buyers"`  // 搶票人數
	SuccessCount int          `json:"success_count"` // 成功幾個
	FailCount   int           `json:"fail_count"`    // 失敗幾個
	StockBefore int           `json:"stock_before"`  // 搶票前庫存
	StockAfter  int           `json:"stock_after"`   // 搶票後庫存
	Results     []BuyerResult `json:"results"`       // 每個人的結果
	DurationMs  int64         `json:"duration_ms"`   // 花了幾毫秒
	Explanation string        `json:"explanation"`   // 白話說明
}

// ============================================================
// 版本一：不加鎖的搶票（會超賣！）
// ============================================================
//
// 每個 goroutine 的流程：
//   Step 1: 查庫存
//   Step 2: 如果 stock <= 0 → 失敗
//   Step 3: 扣庫存
//
// 問題在哪？
//   Step 1 和 Step 3 之間沒有鎖，所以可能發生：
//     goroutine A: 查庫存 → 看到 stock = 5
//     goroutine B: 查庫存 → 也看到 stock = 5（A 還沒扣！）
//     goroutine A: 扣庫存 → stock 變 4
//     goroutine B: 扣庫存 → stock 也變 4（應該要變 3 的！）
//   → 兩個人都買到了，但庫存只扣了 1 → 超賣！

func RunRushWithoutLock(req RushRequest) RushResult {
	start := time.Now()

	// 先記錄搶票前的庫存
	var stockBefore int
	database.DB.QueryRow("SELECT stock FROM events WHERE id = $1", req.EventID).Scan(&stockBefore)

	// TODO: 宣告 results，用 make([]BuyerResult, req.Buyers) 預先分配好位置
	//       每個 goroutine 寫自己對應的 index（results[buyerID]），這樣不同 goroutine 不會互相干擾
	results := make([]BuyerResult, req.Buyers)

	// TODO: 宣告 var wg sync.WaitGroup
	var wg sync.WaitGroup

	// TODO: 用 for 迴圈跑 req.Buyers 次（i 從 0 到 req.Buyers-1）
	//
	//   每一輪要做的事：
	//     1. wg.Add(1)
	//
	//     2. go func(buyerID int) { ... }(i)  — 把 i 當參數傳進去，避免閉包陷阱
	//
	//     goroutine 裡面要做的事：
	//       a. defer wg.Done()
	//
	//       b. 用 database.DB.QueryRow 查該活動的庫存
	//          查詢失敗 → results[buyerID] = BuyerResult{BuyerID: buyerID+1, Success: false, Message: "查詢失敗: " + err.Error()}
	//          然後 return
	//
	//       c. 判斷庫存：如果 stock <= 0
	//          → results[buyerID] = BuyerResult{BuyerID: buyerID+1, Success: false, Message: "票已售完"}
	//          然後 return
	//
	//       d. 用 database.DB.Exec 扣庫存（stock - 1），庫存歸零時把 available 設為 false
	//          ⚠️ 這裡沒有鎖！查詢和更新之間其他 goroutine 可能也在做同樣的事 → race condition
	//          失敗 → 設定失敗結果，return
	//
	//       e. 成功 → results[buyerID] = BuyerResult{BuyerID: buyerID+1, Success: true, Message: "搶票成功！"}

	for i := 0; i < req.Buyers; i++ {
		wg.Add(1)

		go func (buyerID int) {
			defer wg.Done() // 離開時告訴 WaitGroup：這名買家忙完了

			var stock int
			err := database.DB.QueryRow("SELECT stock FROM events WHERE id = $1", req.EventID).Scan(&stock)
			if err != nil {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "查詢失敗: " + err.Error()}
				return
			}

			if stock <= 0 {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "票已售完"}
				return
			}

			_, err = database.DB.Exec("UPDATE events SET stock = stock - 1, available = (stock - 1 > 0) WHERE id = $1", req.EventID)
			if err != nil {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "扣庫存失敗: " + err.Error()}
				return
			}

			results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: true, Message: "搶票成功！"}
		}(i)
	}
	// TODO: wg.Wait()  — 等所有 goroutine 都跑完
	wg.Wait()

	// 搶票後查庫存
	var stockAfter int
	database.DB.QueryRow("SELECT stock FROM events WHERE id = $1", req.EventID).Scan(&stockAfter)

	// 統計成功/失敗
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	duration := time.Since(start).Milliseconds()

	return RushResult{
		Type:         "搶票模擬（沒有加鎖 ⚠️）",
		EventID:      req.EventID,
		TotalBuyers:  req.Buyers,
		SuccessCount: successCount,
		FailCount:    failCount,
		StockBefore:  stockBefore,
		StockAfter:   stockAfter,
		Results:      results,
		DurationMs:   duration,
		Explanation: fmt.Sprintf(
			"沒有加鎖！%d 人同時搶票，庫存從 %d 變成 %d（賣出 %d 張），但有 %d 人說搶到了。如果「賣出數量 < 成功人數」就代表超賣了！",
			req.Buyers, stockBefore, stockAfter, stockBefore-stockAfter, successCount,
		),
	}
}

// ============================================================
// 版本二：加 SELECT FOR UPDATE 鎖的搶票（正確！）
// ============================================================
//
// 跟版本一的差異：
//   每個 goroutine 都在 transaction 裡面用 SELECT ... FOR UPDATE
//   這樣每次只有一個 goroutine 能「查庫存 + 扣庫存」，所以不會超賣

func RunRushWithLock(req RushRequest) RushResult {
	start := time.Now()

	// 搶票前庫存
	var stockBefore int
	database.DB.QueryRow("SELECT stock FROM events WHERE id = $1", req.EventID).Scan(&stockBefore)

	// TODO: 宣告 results，用 make([]BuyerResult, req.Buyers) 預先分配
	results := make([]BuyerResult, req.Buyers)

	// TODO: 宣告 var wg sync.WaitGroup
	var wg sync.WaitGroup

	// TODO: 用 for 迴圈跑 req.Buyers 次
	//
	//   每一輪要做的事：
	//     1. wg.Add(1)
	//
	//     2. go func(buyerID int) { ... }(i)
	//
	//     goroutine 裡面要做的事（注意跟版本一的差異 ✅）：
	//       a. defer wg.Done()
	//
	//       b. ✅ 開一個 transaction（database.DB.Begin()）
	//          失敗 → 設定失敗結果，return
	//          defer tx.Rollback()  — 中途出錯會自動回滾，也會釋放鎖
	//
	//       c. ✅ 用 tx.QueryRow 查庫存，SQL 最後要加上 FOR UPDATE
	//          注意：用 tx.QueryRow 而不是 database.DB.QueryRow
	//          因為 FOR UPDATE 必須在 transaction 裡面才有效
	//          其他 goroutine 如果也在對同一行做 FOR UPDATE，會在這裡排隊等待
	//
	//       d. 檢查庫存：如果 stock <= 0
	//          → 失敗結果，return（defer tx.Rollback() 會自動釋放鎖）
	//
	//       e. ✅ 用 tx.Exec 扣庫存（stock - 1），庫存歸零時把 available 設為 false
	//          注意：用 tx.Exec 而不是 database.DB.Exec，要在同一個 transaction 裡面
	//
	//       f. ✅ tx.Commit() — 提交 transaction，同時釋放鎖
	//          下一個排隊的 goroutine 這時候才能拿到鎖繼續
	//
	//       g. 成功 → results[buyerID] = BuyerResult{BuyerID: buyerID+1, Success: true, Message: "搶票成功！"}

	for i := 0; i < req.Buyers; i++ {
		wg.Add(1)

		go func (buyerID int) {
			defer wg.Done()

			tx, err := database.DB.Begin()
			if err != nil {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "開啟交易失敗: " + err.Error()}
				return
			}
			defer tx.Rollback() // 中途出錯會自動回滾，釋放鎖

			var stock int
			err = tx.QueryRow("SELECT stock FROM events WHERE id = $1 FOR UPDATE", req.EventID).Scan(&stock)
			if err != nil {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "查詢失敗: " + err.Error()}
				return
			} //  增訂單 (tx.QueryRow)

			if stock <= 0 {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "票已售完"}
				return
			}

			_, err = tx.Exec("UPDATE events SET stock = stock - 1, available = (stock - 1 > 0) WHERE id = $1", req.EventID)
			if err != nil {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "扣庫存失敗: " + err.Error()}
				return
			} // 扣除庫存 (tx.Exec)

			if err := tx.Commit(); err != nil {
				results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: false, Message: "提交交易失敗: " + err.Error()}
				return
			}

			results[buyerID] = BuyerResult{BuyerID: buyerID + 1, Success: true, Message: "搶票成功！"}

		}(i)
		
	}

	// TODO: wg.Wait()
	wg.Wait()

	
	// 搶票後庫存
	var stockAfter int
	database.DB.QueryRow("SELECT stock FROM events WHERE id = $1", req.EventID).Scan(&stockAfter)

	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	duration := time.Since(start).Milliseconds()

	return RushResult{
		Type:         "搶票模擬（有加 SELECT FOR UPDATE 鎖 ✅）",
		EventID:      req.EventID,
		TotalBuyers:  req.Buyers,
		SuccessCount: successCount,
		FailCount:    failCount,
		StockBefore:  stockBefore,
		StockAfter:   stockAfter,
		Results:      results,
		DurationMs:   duration,
		Explanation: fmt.Sprintf(
			"有加鎖！%d 人同時搶票，庫存從 %d 變成 %d（賣出 %d 張），成功 %d 人。「賣出數量 = 成功人數」代表沒有超賣！",
			req.Buyers, stockBefore, stockAfter, stockBefore-stockAfter, successCount,
		),
	}
}
