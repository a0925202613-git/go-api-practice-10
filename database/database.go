package database

import (
	"database/sql"

	"go-api-practice-10/config"

	_ "github.com/lib/pq"
)

var DB *sql.DB

const createTablesSQL = `
CREATE TABLE IF NOT EXISTS events (
	id          SERIAL PRIMARY KEY,
	name        VARCHAR(200) NOT NULL,
	venue       VARCHAR(200) NOT NULL,
	price       INTEGER      NOT NULL,
	total_stock INTEGER      NOT NULL DEFAULT 100,
	stock       INTEGER      NOT NULL DEFAULT 100,
	available   BOOLEAN      NOT NULL DEFAULT true,
	event_date  TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_orders (
	id            SERIAL PRIMARY KEY,
	event_id      INTEGER      NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
	customer_name VARCHAR(255) NOT NULL,
	quantity      INTEGER      NOT NULL DEFAULT 1,
	total_price   INTEGER      NOT NULL,
	status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
	ordered_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	cancelled_at  TIMESTAMP WITH TIME ZONE,
	created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

const seedEventsSQL = `
DO $$
BEGIN
  IF (SELECT COUNT(*) FROM events) = 0 THEN
    INSERT INTO events (name, venue, price, total_stock, stock, available, event_date) VALUES
      ('五月天 2026 巡迴演唱會',     '台北小巨蛋',   2800, 100, 100, true,  '2026-07-15 19:00:00+08'),
      ('周杰倫 嘉年華世界巡迴',      '高雄巨蛋',     3200, 80,  80,  true,  '2026-08-20 19:30:00+08'),
      ('韋禮安 Emo 小巡迴',          '台北 Legacy',  1200, 50,  50,  true,  '2026-06-01 20:00:00+08'),
      ('台北爵士音樂節',             '大安森林公園',    500, 200, 200, true,  '2026-09-10 15:00:00+08'),
      ('落日飛車 Summer Tour',       '台中 TADA',    1500, 30,  30,  true,  '2026-07-01 19:00:00+08'),
      ('已售完的測試活動',            '測試場地',      100,  20,   0, false, '2026-12-25 18:00:00+08');
  END IF;
END $$;
`

func Connect() error {
	var err error
	DB, err = sql.Open("postgres", config.DatabaseURL())
	if err != nil {
		return err
	}
	if err := DB.Ping(); err != nil {
		return err
	}
	if _, err = DB.Exec(createTablesSQL); err != nil {
		return err
	}
	_, err = DB.Exec(seedEventsSQL)
	return err
}
