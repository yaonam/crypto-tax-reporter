package wallet

import "gorm.io/gorm"

func Import(db *gorm.DB, address string)

func getTransfers()

func getTxsFromTransfers()
