package wallet

import (
	"testing"

	"github.com/akmalsulaymonov/alif-wallet/pkg/types"
)

func TestService_FindAccountByID_Success(t *testing.T) {
	// Инициализация сервиса
	s := &Service{}

	// Регистрация аккаунта
	account, err := s.RegisterAccount("12345")
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	// Поиск существующего аккаунта
	foundAccount, err := s.FindAccountByID(account.ID)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Проверка, что указатель на аккаунт совпадает
	if foundAccount != account {
		t.Errorf("expected account %v, got %v", account, foundAccount)
	}
}

func TestService_FindAccountByID_NotFound(t *testing.T) {
	// Инициализация сервиса
	s := &Service{}

	// Поиск несуществующего аккаунта
	_, err := s.FindAccountByID(999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Проверка, что возвращена правильная ошибка
	if err != ErrAccountNotFound {
		t.Errorf("expected error %v, got %v", ErrAccountNotFound, err)
	}
}

func TestService_Reject_Success(t *testing.T) {
	// Инициализация сервиса
	s := &Service{}

	// register user
	account, err := s.RegisterAccount("12345")
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	// Депозит денег
	err = s.Deposit(account.ID, 1000)
	if err != nil {
		t.Fatalf("failed to deposit money: %v", err)
	}

	// Создаём платёж
	payment, err := s.Pay(account.ID, 500, "food")
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// Отменяем платёж
	err = s.Reject(payment.ID)
	if err != nil {
		t.Fatalf("failed to reject payment: %v", err)
	}

	// Проверяем статус платежа
	if payment.Status != types.PaymentStatusFail {
		t.Errorf("expected payment status %v, got %v", types.PaymentStatusFail, payment.Status)
	}

	// Проверяем баланс аккаунта
	if account.Balance != 1000 {
		t.Errorf("expected account balance %v, got %v", 1000, account.Balance)
	}
}

func TestService_Reject_NotFound(t *testing.T) {
	// Инициализация сервиса
	s := &Service{}

	// Попытка отменить несуществующий платёж
	err := s.Reject("nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Проверяем, что вернулась правильная ошибка
	if err != ErrPaymentNotFound {
		t.Errorf("expected error %v, got %v", ErrPaymentNotFound, err)
	}
}

func TestService_Repeat_Success(t *testing.T) {
	// Инициализация сервиса
	s := &Service{}

	// Регистрация аккаунта
	account, err := s.RegisterAccount("12345")
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	// Депозит денег
	err = s.Deposit(account.ID, 1000)
	if err != nil {
		t.Fatalf("failed to deposit money: %v", err)
	}

	// Создание платежа
	originalPayment, err := s.Pay(account.ID, 500, "food")
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	// Повторение платежа
	repeatedPayment, err := s.Repeat(originalPayment.ID)
	if err != nil {
		t.Fatalf("failed to repeat payment: %v", err)
	}

	// Проверка данных повторного платежа
	if repeatedPayment.Amount != originalPayment.Amount {
		t.Errorf("expected amount %v, got %v", originalPayment.Amount, repeatedPayment.Amount)
	}
	if repeatedPayment.Category != originalPayment.Category {
		t.Errorf("expected category %v, got %v", originalPayment.Category, repeatedPayment.Category)
	}
	if repeatedPayment.AccountID != originalPayment.AccountID {
		t.Errorf("expected account ID %v, got %v", originalPayment.AccountID, repeatedPayment.AccountID)
	}
}

func TestService_Repeat_NotFound(t *testing.T) {
	// Инициализация сервиса
	s := &Service{}

	// Попытка повторить несуществующий платёж
	_, err := s.Repeat("nonexistent-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Проверка, что вернулась правильная ошибка
	if err != ErrPaymentNotFound {
		t.Errorf("expected error %v, got %v", ErrPaymentNotFound, err)
	}
}

func TestService_FavoritePayment(t *testing.T) {
	s := &Service{}

	account, err := s.RegisterAccount("12345")
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	err = s.Deposit(account.ID, 1000)
	if err != nil {
		t.Fatalf("failed to deposit money: %v", err)
	}

	payment, err := s.Pay(account.ID, 500, "food")
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	favorite, err := s.FavoritePayment(payment.ID, "My Favorite Payment")
	if err != nil {
		t.Fatalf("failed to create favorite: %v", err)
	}

	if favorite.Name != "My Favorite Payment" {
		t.Errorf("expected favorite name %v, got %v", "My Favorite Payment", favorite.Name)
	}
}

func TestService_PayFromFavorite(t *testing.T) {
	s := &Service{}

	account, err := s.RegisterAccount("12345")
	if err != nil {
		t.Fatalf("failed to register account: %v", err)
	}

	err = s.Deposit(account.ID, 1000)
	if err != nil {
		t.Fatalf("failed to deposit money: %v", err)
	}

	payment, err := s.Pay(account.ID, 500, "food")
	if err != nil {
		t.Fatalf("failed to create payment: %v", err)
	}

	favorite, err := s.FavoritePayment(payment.ID, "Favorite")
	if err != nil {
		t.Fatalf("failed to create favorite: %v", err)
	}

	newPayment, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Fatalf("failed to pay from favorite: %v", err)
	}

	if newPayment.Amount != favorite.Amount {
		t.Errorf("expected payment amount %v, got %v", favorite.Amount, newPayment.Amount)
	}
}

func TestService_Export_success_user(t *testing.T) {
	s := &Service{}

	s.RegisterAccount("+992000000001")
	s.RegisterAccount("+992000000002")
	s.RegisterAccount("+992000000003")

	err := s.ExportToFile("export.txt")
	if err != nil {
		t.Errorf("method ExportToFile returned not nil error, err => %v", err)
	}

}

func TestService_Import_success_user(t *testing.T) {
	s := &Service{}

	err := s.ExportToFile("export.txt")
	if err != nil {
		t.Errorf("method ExportToFile returned not nil error, err => %v", err)
	}

	err = s.ImportFromFile("export.txt")

	if err != nil {
		t.Errorf("method ImportToFile returned not nil error, err => %v", err)
	}

}
