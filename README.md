# Practice 10 — 活動售票系統 🎫

延續 practice-9 的架構，加入 **goroutine 搶票模擬** 和 **database lock（SELECT FOR UPDATE）** 練習。

## 練習主題

| 分類 | 練習內容 |
|------|----------|
| CRUD API | 活動（events）的新增、查詢、更新、刪除 |
| 訂單系統 | 購票、取消訂單（含 transaction 和庫存管理） |
| Goroutine | 多個 goroutine 同時搶票，用 sync.WaitGroup 協調 |
| Database Lock | `SELECT ... FOR UPDATE` 防止超賣（race condition） |

## 啟動方式

```bash
# 1. 建立 PostgreSQL 資料庫
createdb practice10

# 2. 確認 .env 的 DATABASE_URL 正確

# 3. 啟動
go run main.go
# Server running on http://localhost:8085
```

## API 列表

### 公開 API
| Method | Endpoint | 說明 |
|--------|----------|------|
| GET | /api/events | 活動列表（可加 `?available=true`) |
| GET | /api/events/:id | 單一活動 |
| GET | /api/ticket-orders | 訂單列表 |
| GET | /api/ticket-orders/:id | 單一訂單 |

### 需 Token（`Authorization: Bearer demo-token-123`）
| Method | Endpoint | 說明 |
|--------|----------|------|
| POST | /api/events | 新增活動 |
| PUT | /api/events/:id | 更新活動 |
| DELETE | /api/events/:id | 刪除活動 |
| POST | /api/ticket-orders | 購票 |
| POST | /api/ticket-orders/:id/cancel | 取消訂單（還原庫存） |

### 搶票模擬（需 Token）
| Method | Endpoint | 說明 |
|--------|----------|------|
| POST | /api/rush/without-lock | 不加鎖搶票（觀察超賣） |
| POST | /api/rush/with-lock | 加 SELECT FOR UPDATE 鎖搶票（不會超賣） |

## 搶票模擬怎麼玩

### 測試超賣（不加鎖）
```bash
curl -X POST http://localhost:8085/api/rush/without-lock \
  -H "Authorization: Bearer demo-token-123" \
  -H "Content-Type: application/json" \
  -d '{"event_id": 5, "buyers": 20}'
```
> 活動 5（落日飛車）只有 30 張票，20 人同時搶。
> 觀察 `success_count` 和實際 `stock_before - stock_after` 是否一致。
> 如果不一致 → 超賣了！

### 測試正確搶票（有加鎖）
```bash
curl -X POST http://localhost:8085/api/rush/with-lock \
  -H "Authorization: Bearer demo-token-123" \
  -H "Content-Type: application/json" \
  -d '{"event_id": 4, "buyers": 20}'
```
> 活動 4（台北爵士音樂節）有 200 張票，20 人同時搶。
> `success_count` 一定等於 `stock_before - stock_after`，不會超賣！

## 核心觀念

### Race Condition（競爭條件）
多個 goroutine 同時讀寫同一筆資料，導致結果不正確。

### SELECT ... FOR UPDATE
在 transaction 中鎖住特定資料行，確保「查庫存 → 扣庫存」是原子操作：
```sql
BEGIN;
SELECT stock FROM events WHERE id = 1 FOR UPDATE;  -- 鎖住
UPDATE events SET stock = stock - 1 WHERE id = 1;
COMMIT;  -- 解鎖
```

### sync.WaitGroup
協調多個 goroutine，等所有人完成後才繼續：
```go
var wg sync.WaitGroup
wg.Add(1)           // 有人要忙
go func() {
    defer wg.Done() // 忙完了
}()
wg.Wait()           // 等所有人忙完
```
