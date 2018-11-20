package migrations

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
	"time"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			cr := &model.Currency{}
			if err := db.C(manager.TableCurrency).Find(bson.M{"code_a3": "EUR"}).One(&cr); err != nil {
				return err
			}

			ps := &model.PaymentSystem{}
			if err := db.C(manager.TablePaymentSystem).Find(bson.M{"name": "CardPay"}).One(ps); err != nil {
				return err
			}

			pms := []interface{}{
				&model.PaymentMethod{
					Id: bson.NewObjectId(),
					Name: "Qiwi",
					PaymentSystem: ps,
					Currency: cr,
					GroupAlias: "qiwi",
					MinPaymentAmount: 0.01,
					MaxPaymentAmount: 15000.00,
					IsActive: true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				&model.PaymentMethod{
					Id: bson.NewObjectId(),
					Name: "WebMoney",
					PaymentSystem: ps,
					Currency: cr,
					GroupAlias: "webmoney",
					MinPaymentAmount: 0.01,
					MaxPaymentAmount: 15000.00,
					IsActive: true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				&model.PaymentMethod{
					Id: bson.NewObjectId(),
					Name: "Neteller",
					PaymentSystem: ps,
					Currency: cr,
					GroupAlias: "neteller",
					MinPaymentAmount: 0.01,
					MaxPaymentAmount: 15000.00,
					IsActive: true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			}

			return db.C(manager.TablePaymentMethod).Insert(pms...)
		},
		func(db *mgo.Database) error {
			var pms []*model.PaymentMethod

			err := db.C(manager.TablePaymentMethod).Find(bson.M{"name": bson.M{"$in": []string{"Qiwi", "WebMoney", "Neteller"}}}).All(&pms)

			if err != nil {
				return err
			}

			if len(pms) < 3 {
				return errors.New("payment methods not found")
			}

			return db.C(manager.TablePaymentMethod).Remove(pms)
		},
	)

	if err != nil {
		return
	}
}
