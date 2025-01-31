package wallet

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

func (s *Service) ImportFromFile(path string) error {

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	str := strings.Split(string(data), "|")

	for i := 0; i < len(str)-1; i++ {
		str_item := strings.Split(str[i], ";")
		id, _ := strconv.ParseInt(str_item[0], 10, 64)
		balance, _ := strconv.ParseInt(str_item[2], 10, 64)
		phone := (str_item[1])

		s.accounts = append(s.accounts, &types.Account{
			ID:      id,
			Phone:   types.Phone(phone),
			Balance: types.Money(balance),
		})
	}

	return nil
}

// Метод Export сохраняет данные accounts, payments и favorites в файлы, если они существуют.
func (s *Service) Export(dir string) error {
	// Экспорт аккаунтов
	if len(s.accounts) > 0 {
		file, err := os.Create(filepath.Join(dir, "accounts.dump"))
		if err != nil {
			return err
		}
		defer file.Close()

		for _, account := range s.accounts {
			_, err := fmt.Fprintf(file, "%d;%s;%d\n", account.ID, account.Phone, account.Balance)
			if err != nil {
				return err
			}
		}
	}

	// Экспорт платежей
	if len(s.payments) > 0 {
		file, err := os.Create(filepath.Join(dir, "payments.dump"))
		if err != nil {
			return err
		}
		defer file.Close()

		for _, payment := range s.payments {
			_, err := fmt.Fprintf(file, "%s;%d;%d;%s;%s\n", payment.ID, payment.AccountID, payment.Amount, payment.Category, payment.Status)
			if err != nil {
				return err
			}
		}
	}

	// Экспорт избранного
	if len(s.favorites) > 0 {
		file, err := os.Create(filepath.Join(dir, "favorites.dump"))
		if err != nil {
			return err
		}
		defer file.Close()

		for _, favorite := range s.favorites {
			_, err := fmt.Fprintf(file, "%s;%d;%s;%d;%s\n", favorite.ID, favorite.AccountID, favorite.Name, favorite.Amount, favorite.Category)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Метод Import загружает данные из файлов, обновляя существующие записи и добавляя новые.
func (s *Service) Import(dir string) error {
	// Импорт аккаунтов
	file, err := os.Open(filepath.Join(dir, "accounts.dump"))
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			var id int64
			var phone types.Phone
			var balance types.Money

			parts := strings.Split(scanner.Text(), ";")
			if len(parts) != 3 {
				return errors.New("invalid accounts file format")
			}

			id, _ = strconv.ParseInt(parts[0], 10, 64)
			phone = types.Phone(parts[1])
			val, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return err // обработка ошибки
			}
			balance = types.Money(val) // Явное приведение типа

			account, err := s.FindAccountByID(id)
			if err == ErrAccountNotFound {
				s.accounts = append(s.accounts, &types.Account{ID: id, Phone: phone, Balance: balance})
			} else {
				account.Balance = balance
			}

			if id > s.nextAccountID {
				s.nextAccountID = id
			}
		}
	}

	// Импорт платежей
	file, err = os.Open(filepath.Join(dir, "payments.dump"))
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			var id string
			var accountID int64
			var amount types.Money
			var category types.PaymentCategory
			var status types.PaymentStatus

			parts := strings.Split(scanner.Text(), ";")
			if len(parts) != 5 {
				return errors.New("invalid payments file format")
			}

			id = parts[0]
			accountID, _ = strconv.ParseInt(parts[1], 10, 64)

			val, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return err // обработка ошибки
			}
			amount = types.Money(val) // Явное приведение типа

			category = types.PaymentCategory(parts[3])
			status = types.PaymentStatus(parts[4])

			s.payments = append(s.payments, &types.Payment{
				ID:        id,
				AccountID: accountID,
				Amount:    amount,
				Category:  category,
				Status:    status,
			})
		}
	}

	// Импорт избранного
	file, err = os.Open(filepath.Join(dir, "favorites.dump"))
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			var id string
			var accountID int64
			var name string
			var amount types.Money
			var category types.PaymentCategory

			parts := strings.Split(scanner.Text(), ";")
			if len(parts) != 5 {
				return errors.New("invalid favorites file format")
			}

			id = parts[0]
			accountID, _ = strconv.ParseInt(parts[1], 10, 64)
			name = parts[2]

			val, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return err // обработка ошибки
			}
			amount = types.Money(val) // Явное приведение типа

			category = types.PaymentCategory(parts[4])

			s.favorites = append(s.favorites, &types.Favorite{
				ID:        id,
				AccountID: accountID,
				Name:      name,
				Amount:    amount,
				Category:  category,
			})
		}
	}

	return nil
}

// Этот метод получает историю платежей конкретного аккаунта.
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	account, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	var history []types.Payment
	for _, payment := range s.payments {
		if payment.AccountID == account.ID {
			history = append(history, *payment)
		}
	}

	return history, nil
}

// Метод сохраняет историю платежей в файлы с разделением на части.
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) == 0 {
		return nil
	}

	fileCount := 1
	var file *os.File
	var err error
	var writer *bufio.Writer

	for i, payment := range payments {
		if i%records == 0 {
			if file != nil {
				writer.Flush()
				file.Close()
			}

			filename := filepath.Join(dir, fmt.Sprintf("payments%d.dump", fileCount))
			if fileCount == 1 && len(payments) <= records {
				filename = filepath.Join(dir, "payments.dump")
			}

			file, err = os.Create(filename)
			if err != nil {
				return err
			}
			writer = bufio.NewWriter(file)
			fileCount++
		}

		_, err := writer.WriteString(fmt.Sprintf("%s;%d;%d;%s;%s\n",
			payment.ID, payment.AccountID, payment.Amount, payment.Category, payment.Status))
		if err != nil {
			return err
		}
	}

	if file != nil {
		writer.Flush()
		file.Close()
	}

	return nil
}
