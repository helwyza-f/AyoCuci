package service

// import (
//     "errors"
//     "fmt"
//     "BackendFramework/internal/model"
//     "gorm.io/gorm"
// )

// type CustomerService struct {
//     DB *gorm.DB
// }

// func NewCustomerService(db *gorm.DB) *CustomerService {
//     return &CustomerService{DB: db}
// }

// // GetAllCustomers - Mendapatkan semua customer berdasarkan outlet
// func (s *CustomerService) GetAllCustomers(outletID uint) ([]model.CustomerResponse, error) {
//     var customers []model.Customer

//     if err := s.DB.Where("cust_outlet = ?", outletID).
//         Order("cust_created DESC").
//         Find(&customers).Error; err != nil {
//         return nil, err
//     }

//     var response []model.CustomerResponse
//     for _, customer := range customers {
//         response = append(response, customer.ToResponse())
//     }

//     return response, nil
// }

// // GetCustomerByID - Mendapatkan customer berdasarkan ID
// func (s *CustomerService) GetCustomerByID(id, outletID uint) (*model.CustomerResponse, error) {
//     var customer model.Customer

//     if err := s.DB.Where("cust_id = ? AND cust_outlet = ?", id, outletID).
//         First(&customer).Error; err != nil {
//         if errors.Is(err, gorm.ErrRecordNotFound) {
//             return nil, errors.New("customer tidak ditemukan")
//         }
//         return nil, err
//     }

//     response := customer.ToResponse()
//     return &response, nil
// }

// // CreateCustomer - Membuat customer baru
// func (s *CustomerService) CreateCustomer(input model.CustomerInput, outletID uint, userUpdate string) (*model.CustomerResponse, error) {
//     // Validasi phone sudah ada di outlet yang sama
//     var existingCustomer model.Customer
//     if err := s.DB.Where("cust_phone = ? AND cust_outlet = ?", input.Phone, outletID).
//         First(&existingCustomer).Error; err == nil {
//         return nil, errors.New("nomor telepon sudah terdaftar")
//     }

//     customer := model.Customer{
//         OutletID:     &outletID,
//         Nama:         input.Nama,
//         Phone:        input.Phone,
//         Alamat:       input.Alamat,
//         Gender:       input.Gender,
//         TanggalLahir: input.TanggalLahir,
//         UserUpdate:   userUpdate,
//     }

//     if err := s.DB.Create(&customer).Error; err != nil {
//         return nil, fmt.Errorf("gagal membuat customer: %v", err)
//     }

//     response := customer.ToResponse()
//     return &response, nil
// }

// // UpdateCustomer - Update customer
// func (s *CustomerService) UpdateCustomer(id, outletID uint, input model.UpdateCustomerInput, userUpdate string) (*model.CustomerResponse, error) {
//     var customer model.Customer

//     if err := s.DB.Where("cust_id = ? AND cust_outlet = ?", id, outletID).
//         First(&customer).Error; err != nil {
//         if errors.Is(err, gorm.ErrRecordNotFound) {
//             return nil, errors.New("customer tidak ditemukan")
//         }
//         return nil, err
//     }

//     // Validasi phone jika diubah
//     if input.Phone != "" && input.Phone != customer.Phone {
//         var existingCustomer model.Customer
//         if err := s.DB.Where("cust_phone = ? AND cust_outlet = ? AND cust_id != ?",
//             input.Phone, outletID, id).
//             First(&existingCustomer).Error; err == nil {
//             return nil, errors.New("nomor telepon sudah digunakan")
//         }
//     }

//     // Update fields
//     if input.Nama != "" {
//         customer.Nama = input.Nama
//     }
//     if input.Phone != "" {
//         customer.Phone = input.Phone
//     }
//     if input.Alamat != "" {
//         customer.Alamat = input.Alamat
//     }
//     if input.Gender != "" {
//         customer.Gender = input.Gender
//     }
//     if input.TanggalLahir != "" {
//         customer.TanggalLahir = input.TanggalLahir
//     }
//     customer.UserUpdate = userUpdate

//     if err := s.DB.Save(&customer).Error; err != nil {
//         return nil, fmt.Errorf("gagal update customer: %v", err)
//     }

//     response := customer.ToResponse()
//     return &response, nil
// }

// // DeleteCustomer - Soft delete customer
// func (s *CustomerService) DeleteCustomer(id, outletID uint) error {
//     var customer model.Customer

//     if err := s.DB.Where("cust_id = ? AND cust_outlet = ?", id, outletID).
//         First(&customer).Error; err != nil {
//         if errors.Is(err, gorm.ErrRecordNotFound) {
//             return errors.New("customer tidak ditemukan")
//         }
//         return err
//     }

//     if err := s.DB.Delete(&customer).Error; err != nil {
//         return fmt.Errorf("gagal menghapus customer: %v", err)
//     }

//     return nil
// }

// // SearchCustomers - Pencarian customer berdasarkan nama
// func (s *CustomerService) SearchCustomers(outletID uint, query string) ([]model.CustomerResponse, error) {
//     var customers []model.Customer

//     if err := s.DB.Where("cust_outlet = ? AND cust_nama LIKE ?", outletID, "%"+query+"%").
//         Order("cust_created DESC").
//         Find(&customers).Error; err != nil {
//         return nil, err
//     }

//     var response []model.CustomerResponse
//     for _, customer := range customers {
//         response = append(response, customer.ToResponse())
//     }

//     return response, nil
// }
