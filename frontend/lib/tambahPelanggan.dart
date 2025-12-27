import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:url_launcher/url_launcher.dart';

class CustomerPage extends StatefulWidget {
  const CustomerPage({super.key});

  @override
  State<CustomerPage> createState() => _CustomerPageState();
}

class _CustomerPageState extends State<CustomerPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  List<Map<String, dynamic>> _customers = [];
  bool _isLoading = false;
  String? _token;
  int? _outletId;
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs
        .getString('access_token')
        ?.replaceAll(RegExp(r'Bearer\s+|\s+'), '');
    _outletId = prefs.getInt('outlet_id');

    if (_token != null && _outletId != null) {
      await _fetchCustomers();
    } else {
      _showSnackBar('Silakan login kembali', Colors.red);
    }
  }

  Future<void> _fetchCustomers() async {
    setState(() => _isLoading = true);

    try {
      final response = await http.get(
        Uri.parse('$_baseUrl/customers?outlet_id=$_outletId'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_token',
        },
      );

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final rawData = data['data'] ?? [];
        
        setState(() {
          _customers = List<Map<String, dynamic>>.from(rawData.map((item) => {
            'id': item['cust_id'],
            'nama': item['cust_nama'],
            'phone': item['cust_phone'],
            'alamat': item['cust_alamat'],
            'gender': item['cust_gender'],
            'tanggal_lahir': item['cust_tanggal_lahir'],
          }));
          _isLoading = false;
        });
      } else {
        setState(() => _isLoading = false);
        _showSnackBar('Gagal memuat data pelanggan', Colors.red);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      _showSnackBar('Terjadi kesalahan: $e', Colors.red);
    }
  }

  Future<void> _saveCustomer({
    int? customerId,
    required String nama,
    required String phone,
    required String alamat,
    required String gender,
    required String tanggalLahir,
  }) async {
    try {
      final url = customerId == null
          ? '$_baseUrl/customers?outlet_id=$_outletId'
          : '$_baseUrl/customers/$customerId?outlet_id=$_outletId';

      final body = {
        'nama': nama,
        'phone': phone,
        'alamat': alamat,
        'gender': gender,
        'tanggal_lahir': tanggalLahir,
      };

      final response = customerId == null
          ? await http.post(
              Uri.parse(url),
              headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer $_token',
              },
              body: json.encode(body),
            )
          : await http.put(
              Uri.parse(url),
              headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer $_token',
              },
              body: json.encode(body),
            );

      if (response.statusCode == 200 || response.statusCode == 201) {
        _showSnackBar(
          customerId == null
              ? 'Pelanggan berhasil ditambahkan'
              : 'Pelanggan berhasil diupdate',
          Colors.green,
        );
        await _fetchCustomers();
      } else {
        final error = json.decode(response.body);
        _showSnackBar(
          error['error'] ?? 'Gagal menyimpan pelanggan',
          Colors.red,
        );
      }
    } catch (e) {
      _showSnackBar('Terjadi kesalahan: $e', Colors.red);
    }
  }

  Future<void> _deleteCustomer(int customerId) async {
    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/customers/$customerId?outlet_id=$_outletId'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_token',
        },
      );

      if (response.statusCode == 200) {
        _showSnackBar('Pelanggan berhasil dihapus', Colors.green);
        await _fetchCustomers();
      } else {
        final error = json.decode(response.body);
        _showSnackBar(
          error['error'] ?? 'Gagal menghapus pelanggan',
          Colors.red,
        );
      }
    } catch (e) {
      _showSnackBar('Terjadi kesalahan: $e', Colors.red);
    }
  }

  Future<void> _checkWhatsApp(String phoneNumber) async {
    if (phoneNumber.isEmpty) {
      _showSnackBar('Nomor WhatsApp tidak boleh kosong', Colors.orange);
      return;
    }

    // Clean phone number
    String cleanPhone = phoneNumber.replaceAll(RegExp(r'[^0-9]'), '');
    
    // Add country code if not exists (Indonesia +62)
    if (cleanPhone.startsWith('0')) {
      cleanPhone = '62${cleanPhone.substring(1)}';
    } else if (!cleanPhone.startsWith('62')) {
      cleanPhone = '62$cleanPhone';
    }

    final whatsappUrl = 'https://wa.me/$cleanPhone';

    try {
      final uri = Uri.parse(whatsappUrl);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        _showSnackBar('Tidak dapat membuka WhatsApp', Colors.red);
      }
    } catch (e) {
      _showSnackBar('Error: $e', Colors.red);
    }
  }

  void _showSnackBar(String message, Color color) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: color,
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  void _showDeleteDialog(int id, String name) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Hapus'),
        content: Text('Hapus pelanggan "$name"?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteCustomer(id);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  void _showCustomerDialog({Map<String, dynamic>? customer}) {
    final isEdit = customer != null;
    final nameCtrl = TextEditingController(text: customer?['nama'] ?? '');
    final phoneCtrl = TextEditingController(text: customer?['phone'] ?? '');
    final addressCtrl = TextEditingController(text: customer?['alamat'] ?? '');
    String selectedGender = customer?['gender'] ?? 'Pria';
    final birthDateCtrl = TextEditingController(text: customer?['tanggal_lahir'] ?? '');

    showDialog(
      context: context,
      barrierColor: Colors.black.withOpacity(0.6),
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => Dialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          insetPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 40),
          backgroundColor: Colors.white,
          child: Container(
            width: double.infinity,
            constraints: BoxConstraints(
              maxHeight: MediaQuery.of(context).size.height * 0.8,
              maxWidth: 400,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.fromLTRB(20, 16, 16, 16),
                  decoration: const BoxDecoration(
                    color: Colors.white,
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Tambah Customer',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Colors.black,
                        ),
                      ),
                      InkWell(
                        onTap: () => Navigator.pop(context),
                        borderRadius: BorderRadius.circular(20),
                        child: const Padding(
                          padding: EdgeInsets.all(4),
                          child: Icon(Icons.close, size: 20, color: Colors.black87),
                        ),
                      ),
                    ],
                  ),
                ),
                // Content
                Flexible(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.fromLTRB(20, 8, 20, 20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Nama Customer
                        _buildLabel('*Nama Customer'),
                        const SizedBox(height: 8),
                        _buildTextField(
                          nameCtrl,
                          'Masukkan Nama Customer',
                          showLabel: false,
                        ),
                        const SizedBox(height: 16),
                        
                        // No WhatsApp
                        _buildLabel('*No Whatsapp (Kirim Nota)'),
                        const SizedBox(height: 8),
                        Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Expanded(
                              child: _buildTextField(
                                phoneCtrl,
                                '0897-2345-6890',
                                showLabel: false,
                              ),
                            ),
                            const SizedBox(width: 8),
                            SizedBox(
                              height: 44,
                              child: ElevatedButton(
                                onPressed: () => _checkWhatsApp(phoneCtrl.text),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: const Color(0xFF4CAF50),
                                  padding: const EdgeInsets.symmetric(horizontal: 20),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  elevation: 0,
                                ),
                                child: const Text(
                                  'Cek',
                                  style: TextStyle(
                                    color: Colors.white,
                                    fontWeight: FontWeight.w600,
                                    fontSize: 13,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        
                        // Alamat
                        _buildLabel('Alamat'),
                        const SizedBox(height: 8),
                        _buildTextField(
                          addressCtrl,
                          'Masukkan Alamat Customer',
                          showLabel: false,
                        ),
                        const SizedBox(height: 16),
                        
                        // Jenis Kelamin
                        _buildLabel('*Jenis Kelamin'),
                        const SizedBox(height: 8),
                        Row(
                          children: [
                            Expanded(
                              child: InkWell(
                                onTap: () {
                                  setDialogState(() => selectedGender = 'Pria');
                                },
                                borderRadius: BorderRadius.circular(4),
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 6),
                                  child: Row(
                                    children: [
                                      Radio<String>(
                                        value: 'Pria',
                                        groupValue: selectedGender,
                                        onChanged: (value) {
                                          setDialogState(() => selectedGender = value!);
                                        },
                                        activeColor: const Color(0xFFFF5252),
                                        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                        visualDensity: const VisualDensity(horizontal: -4, vertical: -4),
                                      ),
                                      const SizedBox(width: 8),
                                      const Text(
                                        'Pria',
                                        style: TextStyle(
                                          fontSize: 13,
                                          color: Colors.black87,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                            Expanded(
                              child: InkWell(
                                onTap: () {
                                  setDialogState(() => selectedGender = 'Wanita');
                                },
                                borderRadius: BorderRadius.circular(4),
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 6),
                                  child: Row(
                                    children: [
                                      Radio<String>(
                                        value: 'Wanita',
                                        groupValue: selectedGender,
                                        onChanged: (value) {
                                          setDialogState(() => selectedGender = value!);
                                        },
                                        activeColor: const Color(0xFFFF5252),
                                        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                                        visualDensity: const VisualDensity(horizontal: -4, vertical: -4),
                                      ),
                                      const SizedBox(width: 8),
                                      const Text(
                                        'Wanita',
                                        style: TextStyle(
                                          fontSize: 13,
                                          color: Colors.black87,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        
                        // Tanggal Lahir
                        _buildLabel('Tanggal Lahir'),
                        const SizedBox(height: 8),
                        _buildTextField(
                          birthDateCtrl,
                          'Tanggal lahir',
                          showLabel: false,
                          readOnly: true,
                          onTap: () async {
                            final date = await showDatePicker(
                              context: context,
                              initialDate: DateTime.now(),
                              firstDate: DateTime(1900),
                              lastDate: DateTime.now(),
                              builder: (context, child) {
                                return Theme(
                                  data: Theme.of(context).copyWith(
                                    colorScheme: const ColorScheme.light(
                                      primary: Color(0xFFFF5252),
                                    ),
                                  ),
                                  child: child!,
                                );
                              },
                            );
                            if (date != null) {
                              birthDateCtrl.text =
                                  '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
                            }
                          },
                        ),
                      ],
                    ),
                  ),
                ),
                // Footer Button
                Padding(
                  padding: const EdgeInsets.fromLTRB(20, 0, 20, 20),
                  child: SizedBox(
                    width: double.infinity,
                    height: 48,
                    child: ElevatedButton(
                      onPressed: () {
                        if (nameCtrl.text.isEmpty || phoneCtrl.text.isEmpty) {
                          _showSnackBar(
                            'Nama dan No Whatsapp wajib diisi',
                            Colors.red,
                          );
                          return;
                        }
                        Navigator.pop(context);
                        _saveCustomer(
                          customerId: customer?['id'],
                          nama: nameCtrl.text,
                          phone: phoneCtrl.text,
                          alamat: addressCtrl.text,
                          gender: selectedGender,
                          tanggalLahir: birthDateCtrl.text,
                        );
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFFF5757),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        'TAMBAH',
                        style: TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.w600,
                          fontSize: 14,
                          letterSpacing: 0.8,
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildLabel(String text) {
    return Text(
      text,
      style: const TextStyle(
        fontSize: 12,
        fontWeight: FontWeight.w400,
        color: Colors.black87,
      ),
    );
  }

  Widget _buildTextField(
    TextEditingController ctrl,
    String hint, {
    bool readOnly = false,
    VoidCallback? onTap,
    bool showLabel = true,
  }) {
    return SizedBox(
      height: 44,
      child: TextField(
        controller: ctrl,
        readOnly: readOnly,
        onTap: onTap,
        style: const TextStyle(fontSize: 13, color: Colors.black87),
        decoration: InputDecoration(
          hintText: hint,
          hintStyle: TextStyle(
            color: Colors.grey[400],
            fontSize: 13,
          ),
          filled: true,
          fillColor: Colors.white,
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 14,
            vertical: 12,
          ),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(color: Colors.grey[300]!, width: 1),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: BorderSide(color: Colors.grey[300]!, width: 1),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(8),
            borderSide: const BorderSide(color: Color(0xFFFF5252), width: 1.5),
          ),
          isDense: true,
        ),
      ),
    );
  }

  List<Map<String, dynamic>> get _filteredCustomers {
    if (_searchQuery.isEmpty) return _customers;
    return _customers
        .where(
          (c) =>
              c['nama']?.toLowerCase().contains(_searchQuery.toLowerCase()) ??
              false,
        )
        .toList();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          'Kembali',
          style: TextStyle(color: Colors.black, fontSize: 16),
        ),
      ),
      body: Column(
        children: [
          // Search Bar
          Container(
            color: Colors.white,
            padding: const EdgeInsets.all(16),
            child: TextField(
              onChanged: (value) => setState(() => _searchQuery = value),
              decoration: InputDecoration(
                hintText: 'Cari Nama Customer',
                hintStyle: TextStyle(color: Colors.grey[400]),
                prefixIcon: Icon(Icons.search, color: Colors.grey[400]),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey[300]!),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey[300]!),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: Color(0xFFFF6B6B)),
                ),
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
              ),
            ),
          ),
          const SizedBox(height: 16),
          // Customer List
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(color: Color(0xFFFF6B6B)),
                  )
                : _filteredCustomers.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.person_outline,
                              size: 80,
                              color: Colors.grey[300],
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'Belum ada data pelanggan',
                              style: TextStyle(color: Colors.grey[600]),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: _filteredCustomers.length,
                        itemBuilder: (context, index) {
                          final customer = _filteredCustomers[index];
                          return Container(
                            margin: const EdgeInsets.only(bottom: 12),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: ListTile(
                              leading: CircleAvatar(
                                backgroundColor: Colors.grey[300],
                                child: Icon(
                                  Icons.person_outline,
                                  color: Colors.grey[600],
                                ),
                              ),
                              title: Text(
                                customer['nama'] ?? '',
                                style: const TextStyle(
                                    fontWeight: FontWeight.w600),
                              ),
                              subtitle: Text(
                                customer['phone'] ?? '',
                                style: TextStyle(
                                  color: Colors.grey[600],
                                  fontSize: 12,
                                ),
                              ),
                              trailing: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  IconButton(
                                    icon: const Icon(
                                      Icons.edit,
                                      color: Color(0xFF4CAF50),
                                    ),
                                    onPressed: () =>
                                        _showCustomerDialog(customer: customer),
                                  ),
                                  IconButton(
                                    icon: const Icon(
                                      Icons.delete,
                                      color: Colors.red,
                                    ),
                                    onPressed: () => _showDeleteDialog(
                                      customer['id'],
                                      customer['nama'],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          );
                        },
                      ),
          ),
          // Bottom Button
          Padding(
            padding: const EdgeInsets.all(16),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => _showCustomerDialog(),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFFF6B6B),
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                child: const Text(
                  'TAMBAH CUSTOMER',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}