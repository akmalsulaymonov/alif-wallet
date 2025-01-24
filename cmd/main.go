package main

import (
	"fmt"

	"github.com/akmalsulaymonov/alif-wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}

	// Register account
	account, err := svc.RegisterAccount("+992918246924")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Аккаунт пользователя", account)

	// Deposit to account
	err = svc.Deposit(account.ID, 10)
	if err != nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма дожлна быть положительной")
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")
		}
		return
	}

	fmt.Println("account after deposit:", account)

	// Make a payment
	payment, err := svc.Pay(account.ID, 5, "food")
	if err != nil {
		switch err {
		case wallet.ErrAmountMustBePositive:
			fmt.Println("Сумма дожлна быть положительной")
		case wallet.ErrAccountNotFound:
			fmt.Println("Аккаунт пользователя не найден")
		case wallet.ErrNotEnoughBalance:
			fmt.Println("Недостаточно баланса")
		}
		return
	}

	fmt.Println("Payment:", payment)
	fmt.Println("balance after payment:", account.Balance)

	// Find account by ID
	account, err = svc.FindAccountByID(1)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Account:", account)

	// Find payment by ID
	payment, err = svc.FindPaymentByID(payment.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Payment:", payment)

	// reject payment
	err = svc.Reject(payment.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Account after rejection:", account)
	fmt.Println("Payment after rejection:", payment)

	// repeat a paeyment
	pt, err := svc.Repeat(payment.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Repeated payment:", pt)

	// favorites
	favorite, err := svc.FavoritePayment(payment.ID, "FooD")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Favorite added:", favorite)

	// pay from favorite
	pf, err := svc.PayFromFavorite(favorite.ID)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Paid from Favorite:", pf)
}
