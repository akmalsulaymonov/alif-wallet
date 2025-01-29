package wallet

import (
	"errors"
	"fmt"
	"os"

	"github.com/akmalsulaymonov/alif-wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be > 0")
var ErrAccountNotFound = errors.New("account not found")
var ErrNotEnoughBalance = errors.New("not enough balance in wallet")
var ErrPaymentNotFound = errors.New("payment not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite // Список избранных платежей
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++

	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount

	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}

	account.Balance -= amount

	paymentID := uuid.New().String()

	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}

	s.payments = append(s.payments, payment)

	return payment, nil

}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {
	// find payment by ID
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return ErrPaymentNotFound
	}

	// find account by ID
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return ErrAccountNotFound
	}

	// return to account
	account.Balance += payment.Amount

	// update payment status
	payment.Status = types.PaymentStatusFail

	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	// find payment by ID
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	// find account by ID
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	result, err := s.Pay(account.ID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	// Находим существующий платёж
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, ErrPaymentNotFound
	}

	// Создаём новый элемент избранного
	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}

	// Добавляем в список избранного
	s.favorites = append(s.favorites, favorite)

	return favorite, nil
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	// Находим элемент избранного
	var favorite *types.Favorite
	for _, f := range s.favorites {
		if f.ID == favoriteID {
			favorite = f
			break
		}
	}
	if favorite == nil {
		return nil, errors.New("favorite not found")
	}

	// Проверяем аккаунт
	account, err := s.FindAccountByID(favorite.AccountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	// Проверяем баланс
	if account.Balance < favorite.Amount {
		return nil, ErrNotEnoughBalance
	}

	// Создаём платёж
	payment := &types.Payment{
		ID:        uuid.New().String(),
		AccountID: favorite.AccountID,
		Amount:    favorite.Amount,
		Category:  favorite.Category,
		Status:    types.PaymentStatusInProgress,
	}

	// Обновляем баланс
	account.Balance -= favorite.Amount

	// Добавляем платёж в список
	s.payments = append(s.payments, payment)

	return payment, nil
}

// Method for export Account to file
func (s *Service) ExportToFile(path string) error {

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var str string
	for _, v := range s.accounts {
		str += fmt.Sprint(v.ID) + ";" + string(v.Phone) + ";" + fmt.Sprint(v.Balance) + "|"
	}
	_, err = file.WriteString(str)

	if err != nil {
		return err
	}

	return nil
}
