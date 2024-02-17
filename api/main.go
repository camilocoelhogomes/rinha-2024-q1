package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"strconv"
	"time"
)

type Transactions struct {
	Balance      int64     `gorm:"column:balance"`
	AccountLimit int64     `gorm:"column:account_limit"`
	Value        int64     `gorm:"column:value"`
	Description  string    `gorm:"column:description"`
	CreateTm     time.Time `gorm:"column:create_tm"`
}

type PostTransactionBody struct {
	Value       int64  `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
}

type Saldo struct {
	Total int64     `json:"total"`
	Date  time.Time `json:"data_extrato"`
	Limit int64     `json:"limite"`
}

type LastTransactions struct {
	Value       int64     `json:"valor"`
	TType       string    `json:"tipo"`
	Description string    `json:"descricao"`
	CreateTm    time.Time `json:"realizada_em"`
}

type GetReturn struct {
	Saldo            Saldo              `json:"saldo"`
	LastTransactions []LastTransactions `json:"ultimas_transacoes"`
}

type PostReturn struct {
	Balance int64 `json:"saldo" gorm:"column:balance"`
	Limit   int64 `json:"limite" gorm:"column:account_limit"`
}

func main() {
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_NAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"))
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})

	app := fiber.New()

	app.Get("/clientes/:id/extrato", func(c *fiber.Ctx) error {
		id, _ := strconv.ParseInt(c.Params("id"), 10, 8)
		var tList []Transactions
		var saldo Saldo
		var lastTransactions []LastTransactions
		err := db.Raw("select a.balance, a.account_limit, t.value, t.description, t.create_tm from transactions t inner join accounts a on a.account_id = t.account_id where a.account_id = ? order by t.id desc limit 10", id).Scan(&tList)
		if err.Error != nil {
			return c.Status(404).SendString("")
		}
		for i, item := range tList {
			if i == 0 {
				saldo = Saldo{
					Total: item.Balance,
					Date:  time.Now(),
					Limit: item.AccountLimit,
				}
			}
			if item.Value > 0 {
				lastTransactions = append(lastTransactions, LastTransactions{
					Value:       item.Value,
					TType:       "c",
					Description: item.Description,
					CreateTm:    item.CreateTm,
				})
			} else {
				lastTransactions = append(lastTransactions, LastTransactions{
					Value:       -1 * item.Value,
					TType:       "d",
					Description: item.Description,
					CreateTm:    item.CreateTm,
				})
			}

		}
		// Return the account as JSON response
		return c.JSON(GetReturn{Saldo: saldo, LastTransactions: lastTransactions})
	})

	app.Post("/clientes/:id/transacoes", func(c *fiber.Ctx) error {
		id, _ := strconv.ParseInt(c.Params("id"), 10, 8)
		body := new(PostTransactionBody)
		err := c.BodyParser(body)
		if err != nil {
			return c.Status(400).SendString("")
		}
		if body.Type == "c" {
			var returnValue PostReturn
			err := db.Raw("select * from create_credit_transaction(?,?,?);", id, body.Description, body.Value).Scan(&returnValue)
			if err.Error != nil {
				return c.Status(404).SendString("")
			}
			return c.JSON(returnValue)
		}

		if body.Type == "d" {
			var returnValue PostReturn
			err := db.Raw("select * from create_debit_transaction(?,?,?);", id, body.Description, body.Value).Scan(&returnValue)

			if err.Error != nil {
				if err.Error.Error() == "ERROR: Account Not Found (SQLSTATE P0001)" {
					return c.Status(404).SendString("")
				}
				return c.Status(422).SendString("")
			}
			return c.JSON(returnValue)
		}

		return c.Status(400).SendString("")
	})

	err = app.Listen(fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")))
	if err != nil {
		return
	}
}
