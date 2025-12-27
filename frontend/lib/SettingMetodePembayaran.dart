import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class PaymentMethodPage extends StatefulWidget {
  final int outletId;

  const PaymentMethodPage({super.key, required this.outletId});

  @override
  State<PaymentMethodPage> createState() => _PaymentMethodPageState();
}

class _PaymentMethodPageState extends State<PaymentMethodPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  String? _accessToken;
  bool _isLoading = false;
  String _filterStatus = 'Semua';
  bool _showOnlyActive = false;
  bool _showOnlyInactive = false;

  List<PaymentMethod> _paymentMethods = [];

  List<PaymentMethod> get _displayList {
    if (_filterStatus == 'Aktif') {
      return _paymentMethods.where((m) => m.isActive).toList();
    } else if (_filterStatus == 'Tidak Aktif') {
      return _paymentMethods.where((m) => !m.isActive).toList();
    }
    return _paymentMethods;
  }

  static const Color primaryColor = Color(0xFFEF5350);
  static const List<String> bankList = [
    'BCA',
    'BRI',
    'BNI',
    'Mandiri',
    'CIMB',
    'Permata',
  ];
  static const List<String> ewalletList = [
    'GoPay',
    'OVO',
    'Dana',
    'LinkAja',
    'ShopeePay',
  ];

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() => _accessToken = prefs.getString('access_token'));
    await _fetchPaymentMethods();
  }

  Future<void> _fetchPaymentMethods() async {
    setState(() => _isLoading = true);
    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/payment-methods/outlet/${widget.outletId}'),
        headers: {
          'Authorization': 'Bearer $_accessToken',
          'Content-Type': 'application/json',
        },
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        setState(() {
          _paymentMethods = (data['data'] as List)
              .map((item) => PaymentMethod.fromJson(item))
              .toList();
        });
      } else {
        _showSnackBar('Gagal memuat data: ${response.statusCode}', false);
      }
    } catch (e) {
      _showSnackBar('Error: $e', false);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _savePaymentMethod(
    PaymentMethod? method,
    Map<String, dynamic> data,
  ) async {
    setState(() => _isLoading = true);
    try {
      final url = method == null
          ? '$_baseUrl/payment-methods/outlet/${widget.outletId}'
          : '$_baseUrl/payment-methods/${method.id}?outlet_id=${widget.outletId}';

      print('Request URL: $url');
      print('Request Data: ${json.encode(data)}');

      final response = method == null
          ? await http.post(
              Uri.parse(url),
              headers: {
                'Authorization': 'Bearer $_accessToken',
                'Content-Type': 'application/json',
              },
              body: json.encode(data),
            )
          : await http.put(
              Uri.parse(url),
              headers: {
                'Authorization': 'Bearer $_accessToken',
                'Content-Type': 'application/json',
              },
              body: json.encode(data),
            );

      print('Response Status: ${response.statusCode}');
      print('Response Body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        _showSnackBar(
          'Metode pembayaran berhasil ${method == null ? "ditambahkan" : "diperbarui"}',
          true,
        );
        await _fetchPaymentMethods();
      } else {
        try {
          final error = json.decode(response.body);
          _showSnackBar(error['message'] ?? 'Gagal menyimpan', false);
        } catch (e) {
          _showSnackBar('Gagal menyimpan: ${response.body}', false);
        }
      }
    } catch (e) {
      print('Exception: $e');
      _showSnackBar('Error: $e', false);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _deletePaymentMethod(PaymentMethod method) async {
    setState(() => _isLoading = true);
    try {
      final url =
          '$_baseUrl/payment-methods/${method.id}?outlet_id=${widget.outletId}';

      print('Delete URL: $url');

      final response = await http.delete(
        Uri.parse(url),
        headers: {
          'Authorization': 'Bearer $_accessToken',
          'Content-Type': 'application/json',
        },
      );

      print('Response Status: ${response.statusCode}');
      print('Response Body: ${response.body}');

      if (response.statusCode == 200) {
        _showSnackBar('Metode pembayaran berhasil dihapus', true);
        await _fetchPaymentMethods();
      } else {
        try {
          final error = json.decode(response.body);
          _showSnackBar(error['message'] ?? 'Gagal menghapus', false);
        } catch (e) {
          _showSnackBar('Gagal menghapus: ${response.body}', false);
        }
      }
    } catch (e) {
      print('Exception: $e');
      _showSnackBar('Error: $e', false);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _showSnackBar(String message, bool isSuccess) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(
              isSuccess ? Icons.check_circle : Icons.error_outline,
              color: Colors.white,
            ),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: isSuccess ? Colors.green : Colors.red,
        behavior: SnackBarBehavior.floating,
        duration: Duration(seconds: isSuccess ? 2 : 3),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  void _showDeleteConfirmation(PaymentMethod method) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Text('Hapus Metode Pembayaran'),
        content: Text('Apakah Anda yakin ingin menghapus "${method.name}"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deletePaymentMethod(method);
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(10),
              ),
            ),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  IconData _getCategoryIcon(String? category) {
    switch (category) {
      case 'Transfer':
        return Icons.account_balance;
      case 'E-Wallet':
        return Icons.account_balance_wallet;
      case 'Cash':
        return Icons.money;
      default:
        return Icons.payment;
    }
  }

  Color _getCategoryColor(String? category) {
    switch (category) {
      case 'Transfer':
        return Colors.blue;
      case 'E-Wallet':
        return Colors.purple;
      case 'Cash':
        return Colors.green;
      default:
        return Colors.grey;
    }
  }

  void _showFilterBottomSheet() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
        ),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Fitur Metode Pembayaran',
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: Colors.black87,
              ),
            ),
            const SizedBox(height: 4),
            Text(
              'Fitur Pilihan Metode Pembayaran Tidak Akan Muncul Saat Proses Pembayaran',
              style: TextStyle(fontSize: 12, color: Colors.grey[600]),
            ),
            const SizedBox(height: 24),
            _buildFilterOption(
              'Tidak Aktif',
              !_showOnlyActive && !_showOnlyInactive,
            ),
            const Divider(height: 32),
            _buildExpandableSection(
              'Metode Pembayaran Aktif',
              _showOnlyActive,
              () {
                setState(() {
                  _filterStatus = 'Aktif';
                  _showOnlyActive = true;
                  _showOnlyInactive = false;
                });
                Navigator.pop(context);
              },
            ),
            const Divider(height: 32),
            _buildExpandableSection(
              'Metode Pembayaran Tidak Aktif',
              _showOnlyInactive,
              () {
                setState(() {
                  _filterStatus = 'Tidak Aktif';
                  _showOnlyActive = false;
                  _showOnlyInactive = true;
                });
                Navigator.pop(context);
              },
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildFilterOption(String title, bool isSelected) {
    return InkWell(
      onTap: () {
        setState(() {
          _filterStatus = 'Semua';
          _showOnlyActive = false;
          _showOnlyInactive = false;
        });
        Navigator.pop(context);
      },
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
        decoration: BoxDecoration(
          color: isSelected ? const Color(0xFFFFE5E5) : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              title,
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w500,
                color: isSelected ? Colors.black87 : Colors.black87,
              ),
            ),
            Container(
              width: 24,
              height: 24,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: isSelected ? primaryColor : Colors.grey[300],
              ),
              child: isSelected
                  ? const Icon(Icons.check, size: 16, color: Colors.white)
                  : null,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildExpandableSection(
    String title,
    bool isExpanded,
    VoidCallback onTap,
  ) {
    return InkWell(
      onTap: onTap,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w500,
              color: Colors.black87,
            ),
          ),
          Icon(
            isExpanded ? Icons.expand_less : Icons.expand_more,
            color: Colors.black87,
          ),
        ],
      ),
    );
  }

  void _showPaymentMethodDialog([PaymentMethod? method]) {
    String selectedCategory = method?.category ?? 'Cash';
    bool isActive = method?.isActive ?? true;
    String? selectedBank = method?.bankName;
    String? selectedEwallet = method?.ewalletProvider;

    final accountNumberController = TextEditingController(
      text: method?.accountNumber ?? '',
    );
    final accountHolderController = TextEditingController(
      text: method?.accountHolder ?? '',
    );
    final phoneNumberController = TextEditingController(
      text: method?.phoneNumber ?? '',
    );

    showDialog(
      context: context,
      builder: (context) => Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.symmetric(horizontal: 20),
        child: StatefulBuilder(
          builder: (context, setDialogState) {
            return Container(
              constraints: const BoxConstraints(maxWidth: 400),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(20),
                border: Border.all(color: const Color(0xFFFFA726), width: 2),
              ),
              child: SingleChildScrollView(
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        '${method == null ? "Tambah" : "Edit"} Metode Pembayaran',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                          color: Colors.black87,
                        ),
                      ),
                      const SizedBox(height: 20),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            'Aktif',
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.black87,
                            ),
                          ),
                          Transform.scale(
                            scale: 0.85,
                            child: Switch(
                              value: isActive,
                              onChanged: (value) =>
                                  setDialogState(() => isActive = value),
                              activeColor: primaryColor,
                              activeTrackColor: primaryColor.withOpacity(0.5),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),

                      if (method == null) ...[
                        const Text(
                          'Pilih Kategori Metode Pembayaran',
                          style: TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Container(
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(25),
                            border: Border.all(color: Colors.grey[300]!),
                          ),
                          child: DropdownButtonFormField<String>(
                            value: selectedCategory,
                            decoration: const InputDecoration(
                              border: InputBorder.none,
                              contentPadding: EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 12,
                              ),
                            ),
                            icon: const Icon(Icons.keyboard_arrow_down),
                            items: ['Cash', 'Transfer', 'E-Wallet']
                                .map(
                                  (value) => DropdownMenuItem(
                                    value: value,
                                    child: Text(
                                      value,
                                      style: const TextStyle(fontSize: 14),
                                    ),
                                  ),
                                )
                                .toList(),
                            onChanged: (value) {
                              if (value != null) {
                                setDialogState(() {
                                  selectedCategory = value;
                                  selectedBank = null;
                                  selectedEwallet = null;
                                });
                              }
                            },
                          ),
                        ),
                        const SizedBox(height: 16),
                      ],

                      if (selectedCategory == 'Cash')
                        _buildCashInfo()
                      else if (selectedCategory == 'Transfer') ...[
                        _buildLabel('Pilih Bank'),
                        _buildDropdownField(
                          selectedBank,
                          bankList,
                          'Pilih Bank',
                          (v) => setDialogState(() => selectedBank = v),
                        ),
                        const SizedBox(height: 12),
                        _buildLabel('Nomor Rekening'),
                        _buildTextField(
                          accountNumberController,
                          'Masukkan nomor rekening',
                          isNumber: true,
                        ),
                        const SizedBox(height: 12),
                        _buildLabel('Nama Pemilik'),
                        _buildTextField(
                          accountHolderController,
                          'Nama pemilik rekening',
                        ),
                      ] else if (selectedCategory == 'E-Wallet') ...[
                        _buildLabel('Pilih E-Wallet'),
                        _buildDropdownField(
                          selectedEwallet,
                          ewalletList,
                          'Pilih E-Wallet',
                          (v) => setDialogState(() => selectedEwallet = v),
                        ),
                        const SizedBox(height: 12),
                        _buildLabel('Nomor Telepon'),
                        _buildTextField(
                          phoneNumberController,
                          '08xxxxxxxxxx',
                          isNumber: true,
                        ),
                      ],

                      const SizedBox(height: 24),
                      _buildSaveButton(() {
                        String? errorMsg = _validateInputs(
                          selectedCategory,
                          selectedBank,
                          selectedEwallet,
                          accountNumberController,
                          accountHolderController,
                          phoneNumberController,
                        );

                        if (errorMsg != null) {
                          _showSnackBar(errorMsg, false);
                          return;
                        }

                        Navigator.pop(context);

                        final data = _buildPaymentData(
                          selectedCategory,
                          isActive,
                          selectedBank,
                          selectedEwallet,
                          accountNumberController,
                          accountHolderController,
                          phoneNumberController,
                        );
                        _savePaymentMethod(method, data);
                      }),
                    ],
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  String? _validateInputs(
    String category,
    String? bank,
    String? ewallet,
    TextEditingController accountNum,
    TextEditingController accountHolder,
    TextEditingController phone,
  ) {
    if (category == 'Transfer') {
      if (bank == null) return 'Pilih bank terlebih dahulu';
      if (accountNum.text.isEmpty) return 'Nomor rekening harus diisi';
      if (accountHolder.text.isEmpty)
        return 'Nama pemilik rekening harus diisi';
    } else if (category == 'E-Wallet') {
      if (ewallet == null) return 'Pilih e-wallet terlebih dahulu';
      if (phone.text.isEmpty) return 'Nomor telepon harus diisi';
    }
    return null;
  }

  Map<String, dynamic> _buildPaymentData(
    String category,
    bool isActive,
    String? bank,
    String? ewallet,
    TextEditingController accountNum,
    TextEditingController accountHolder,
    TextEditingController phone,
  ) {
    String name, description;

    if (category == 'Transfer') {
      name = '$bank - Transfer';
      description = '$bank ${accountHolder.text} - ${accountNum.text}';
    } else if (category == 'E-Wallet') {
      name = ewallet!;
      description = phone.text;
    } else {
      name = 'Cash';
      description = 'Pembayaran tunai';
    }

    final Map<String, dynamic> data = {
      'name': name,
      'description': description,
      'category': category,
      'is_active': isActive,
    };

    if (bank != null && bank.isNotEmpty) {
      data['bank_name'] = bank;
    }
    if (accountNum.text.isNotEmpty) {
      data['account_number'] = accountNum.text;
    }
    if (accountHolder.text.isNotEmpty) {
      data['account_holder'] = accountHolder.text;
    }
    if (ewallet != null && ewallet.isNotEmpty) {
      data['ewallet_provider'] = ewallet;
    }
    if (phone.text.isNotEmpty) {
      data['phone_number'] = phone.text;
    }

    return data;
  }

  Widget _buildLabel(String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Text(
        text,
        style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500),
      ),
    );
  }

  Widget _buildTextField(
    TextEditingController controller,
    String hint, {
    bool isNumber = false,
  }) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(25),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: TextField(
        controller: controller,
        keyboardType: isNumber ? TextInputType.number : TextInputType.text,
        decoration: InputDecoration(
          hintText: hint,
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 16,
            vertical: 12,
          ),
        ),
      ),
    );
  }

  Widget _buildDropdownField(
    String? value,
    List<String> items,
    String hint,
    Function(String?) onChanged,
  ) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(25),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: DropdownButtonFormField<String>(
        value: value,
        decoration: const InputDecoration(
          border: InputBorder.none,
          contentPadding: EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        ),
        hint: Text(hint),
        icon: const Icon(Icons.keyboard_arrow_down),
        items: items
            .map(
              (v) => DropdownMenuItem(
                value: v,
                child: Text(v, style: const TextStyle(fontSize: 14)),
              ),
            )
            .toList(),
        onChanged: onChanged,
      ),
    );
  }

  Widget _buildCashInfo() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.green.withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.green.withOpacity(0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.money, color: Colors.green, size: 32),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Pembayaran Tunai',
                  style: TextStyle(
                    fontSize: 15,
                    fontWeight: FontWeight.w600,
                    color: Colors.black87,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  'Metode Cash tidak memerlukan informasi tambahan',
                  style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSaveButton(VoidCallback onPressed) {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton(
        onPressed: onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: primaryColor,
          padding: const EdgeInsets.symmetric(vertical: 14),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(25),
          ),
          elevation: 0,
        ),
        child: const Text(
          'Simpan',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
    );
  }

  Widget _buildPaymentMethodCard(PaymentMethod method) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE0E0E0), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: _getCategoryColor(method.category).withOpacity(0.1),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                _getCategoryIcon(method.category),
                color: _getCategoryColor(method.category),
                size: 24,
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          method.name,
                          style: const TextStyle(
                            fontSize: 15,
                            fontWeight: FontWeight.w600,
                            color: Colors.black87,
                          ),
                        ),
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: method.isActive
                              ? Colors.green.withOpacity(0.1)
                              : Colors.grey.withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          method.isActive ? 'Aktif' : 'Tidak Aktif',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: method.isActive ? Colors.green : Colors.grey,
                          ),
                        ),
                      ),
                    ],
                  ),

                  const SizedBox(height: 6),
                  if (method.category == 'Transfer') ...[
                    if (method.accountNumber != null)
                      Text(
                        'No. Rekening : ${method.accountNumber}',
                        style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                      ),
                    if (method.accountHolder != null)
                      Text(
                        'A/N : ${method.accountHolder}',
                        style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                      ),
                  ]
                  else if (method.category == 'E-Wallet') ...[
                    if (method.phoneNumber != null)
                      Text(
                        'No. HP : ${method.phoneNumber}',
                        style: TextStyle(fontSize: 12, color: Colors.grey[700]),
                      ),
                  ]
                  else ...[
                    Text(
                      'Pembayaran tunai',
                      style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                    ),
                  ],
                ],
              ),
            ),

            const SizedBox(width: 12),
            _buildIconButton(
              Icons.edit_outlined,
              primaryColor,
              () => _showPaymentMethodDialog(method),
            ),
            const SizedBox(width: 8),
            _buildIconButton(
              Icons.delete_outline,
              Colors.white,
              () => _showDeleteConfirmation(method),
              bgColor: Colors.black87,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildIconButton(
    IconData icon,
    Color iconColor,
    VoidCallback onTap, {
    Color? bgColor,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Container(
        width: 36,
        height: 36,
        decoration: BoxDecoration(
          color: bgColor ?? Colors.white,
          shape: BoxShape.circle,
          border: bgColor == null
              ? Border.all(color: primaryColor, width: 2)
              : null,
        ),
        child: Icon(icon, size: 18, color: iconColor),
      ),
    );
  }

  Widget _buildEmptyState() {
    String message;
    if (_filterStatus == 'Aktif') {
      message = 'Belum ada metode pembayaran aktif';
    } else if (_filterStatus == 'Tidak Aktif') {
      message = 'Belum ada metode pembayaran tidak aktif';
    } else {
      message = 'Belum ada metode pembayaran';
    }

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.payment, size: 80, color: Colors.grey[300]),
          const SizedBox(height: 16),
          Text(
            message,
            style: TextStyle(
              fontSize: 16,
              color: Colors.grey[600],
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Tap tombol + untuk menambah metode',
            style: TextStyle(fontSize: 14, color: Colors.grey[400]),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8F9FA),
      appBar: AppBar(
        title: const Text(
          'Metode Pembayaran',
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.bold,
            fontSize: 18,
          ),
        ),
        backgroundColor: Colors.white,
        elevation: 0.5,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: primaryColor))
          : _displayList.isEmpty
          ? _buildEmptyState()
          : ListView(
              padding: const EdgeInsets.fromLTRB(16, 16, 16, 80),
              children: _displayList
                  .map((m) => _buildPaymentMethodCard(m))
                  .toList(),
            ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showPaymentMethodDialog(),
        backgroundColor: primaryColor,
        icon: const Icon(Icons.add, color: Colors.white),
        label: const Text(
          'Tambah Metode',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
        ),
      ),
    );
  }
}

// Model class
class PaymentMethod {
  final int id;
  final String name;
  final String description;
  final String? category;
  final bool isActive;
  final String? bankName;
  final String? accountNumber;
  final String? accountHolder;
  final String? ewalletProvider;
  final String? phoneNumber;

  PaymentMethod({
    required this.id,
    required this.name,
    required this.description,
    this.category,
    required this.isActive,
    this.bankName,
    this.accountNumber,
    this.accountHolder,
    this.ewalletProvider,
    this.phoneNumber,
  });

  factory PaymentMethod.fromJson(Map<String, dynamic> json) {
    return PaymentMethod(
      id: json['id'],
      name: json['name'] ?? '',
      description: json['description'] ?? '',
      category: json['category'],
      isActive: json['is_active'] ?? false,
      bankName: json['bank_name'],
      accountNumber: json['account_number'],
      accountHolder: json['account_holder'],
      ewalletProvider: json['ewallet_provider'],
      phoneNumber: json['phone_number'],
    );
  }
}
