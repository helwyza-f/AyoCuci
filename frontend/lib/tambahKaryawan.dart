import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class KaryawanPage extends StatefulWidget {
  const KaryawanPage({super.key});

  @override
  State<KaryawanPage> createState() => _KaryawanPageState();
}

class _KaryawanPageState extends State<KaryawanPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  List<Map<String, dynamic>> karyawanList = [];
  bool _isLoading = false;
  String? _accessToken;
  int? _outletId;
  Map<int, bool> expandedItems = {};

  final List<String> availablePermissions = [
    "Membuat Order / Transaksi",
    "Menambahkan Order / Transaksi",
    "Membuat Pengaturan",
    "Mengelola Layanan / Produk",
    "Menampilkan Nilai Omzet",
    "Mengelola Data Karyawan",
    "Akses Layanan Transaksi",
    "Akses Layanan Konsep",
    "Akses Layanan Keuangan",
    "Akses Layanan Pelanggan",
  ];

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    final prefs = await SharedPreferences.getInstance();

    String? token = prefs.getString('access_token');
    if (token != null) {
      token = token.trim();
      if (token.startsWith('Bearer ')) {
        token = token.replaceFirst('Bearer ', '').trim();
      }
      token = token.replaceAll('\n', '').replaceAll('\r', '').trim();
    }

    final outletId = prefs.getInt('outlet_id');

    setState(() {
      _accessToken = token;
      _outletId = outletId;
    });

    if (_accessToken != null && _accessToken!.isNotEmpty) {
      if (_outletId != null) {
        await _fetchKaryawanData();
      } else {
        _showSnackBar(
          'Outlet ID tidak ditemukan. Silakan login kembali',
          Colors.red,
        );
      }
    } else {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
    }
  }

  Future<void> _fetchKaryawanData() async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak valid. Silakan login kembali', Colors.red);
      return;
    }

    if (_outletId == null) {
      _showSnackBar('Outlet ID tidak ditemukan', Colors.red);
      return;
    }

    setState(() => _isLoading = true);

    try {
      final url = '$_baseUrl/karyawan?outlet_id=$_outletId';

      final response = await http.get(
        Uri.parse(url),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      if (!mounted) return;

      if (response.statusCode == 200) {
        final data = json.decode(response.body);
        final rawList = List<Map<String, dynamic>>.from(data['data'] ?? []);

        setState(() {
          karyawanList = rawList;
          _isLoading = false;
        });
      } else if (response.statusCode == 401) {
        setState(() => _isLoading = false);
        _showSnackBar('Token tidak valid. Silakan login kembali', Colors.red);
      } else {
        setState(() => _isLoading = false);
        _showSnackBar('Gagal memuat data karyawan', Colors.red);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _saveKaryawan({
    int? karyawanId,
    required String nama,
    required String phone,
    required String email,
    required String password,
    required bool isPremium,
    required List<String> permissions,
  }) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    if (_outletId == null) {
      _showSnackBar('Outlet ID tidak ditemukan', Colors.red);
      return;
    }

    try {
      final url = karyawanId == null
          ? '$_baseUrl/karyawan?outlet_id=$_outletId'
          : '$_baseUrl/karyawan/$karyawanId?outlet_id=$_outletId';

      final body = karyawanId == null
          ? {
              'nama': nama,
              'phone': phone,
              'email': email,
              'password': password,
              'isPremium': isPremium,
              'permissions': permissions,
            }
          : {
              'nama': nama,
              'phone': phone,
              'email': email,
              if (password.isNotEmpty) 'password': password,
              'isPremium': isPremium,
              'permissions': permissions,
            };

      final response = karyawanId == null
          ? await http.post(
              Uri.parse(url),
              headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer $_accessToken',
              },
              body: json.encode(body),
            )
          : await http.put(
              Uri.parse(url),
              headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer $_accessToken',
              },
              body: json.encode(body),
            );

      if (response.statusCode == 200 || response.statusCode == 201) {
        _showSnackBar(
          karyawanId == null
              ? 'Karyawan berhasil ditambahkan'
              : 'Karyawan berhasil diupdate',
          Colors.green,
        );
        await _fetchKaryawanData();
      } else {
        final errorData = json.decode(response.body);
        final errorMessage = errorData['error'] ?? 'Gagal menyimpan karyawan';
        _showSnackBar(errorMessage, Colors.red);
      }
    } catch (e) {
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _deleteKaryawan(int karyawanId) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    if (_outletId == null) {
      _showSnackBar('Outlet ID tidak ditemukan', Colors.red);
      return;
    }

    try {
      final url = '$_baseUrl/karyawan/$karyawanId?outlet_id=$_outletId';

      final response = await http.delete(
        Uri.parse(url),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      if (response.statusCode == 200) {
        _showSnackBar('Karyawan berhasil dihapus', Colors.green);
        await _fetchKaryawanData();
      } else {
        final errorData = json.decode(response.body);
        final errorMessage = errorData['error'] ?? 'Gagal menghapus karyawan';
        _showSnackBar(errorMessage, Colors.red);
      }
    } catch (e) {
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  void _showSnackBar(String message, Color backgroundColor) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: backgroundColor,
          behavior: SnackBarBehavior.floating,
          duration: const Duration(seconds: 3),
        ),
      );
    }
  }

  void _showDeleteConfirmation(int karyawanId, String namaKaryawan) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Hapus'),
        content: Text(
          'Apakah Anda yakin ingin menghapus karyawan "$namaKaryawan"?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteKaryawan(karyawanId);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  void _showTambahKaryawanDialog({Map<String, dynamic>? karyawanData}) {
    final isEdit = karyawanData != null;
    final namaController = TextEditingController(
      text: karyawanData?['kar_nama'] ?? '',
    );
    final phoneController = TextEditingController(
      text: karyawanData?['kar_phone'] ?? '',
    );
    final emailController = TextEditingController(
      text: karyawanData?['kar_email'] ?? '',
    );
    final passwordController = TextEditingController();

    List<String> selectedPermissions = [];
    if (karyawanData != null && karyawanData['kar_permissions'] != null) {
      if (karyawanData['kar_permissions'] is List) {
        selectedPermissions = List<String>.from(
          karyawanData['kar_permissions'],
        );
      } else if (karyawanData['kar_permissions'] is String) {
        try {
          selectedPermissions = List<String>.from(
            json.decode(karyawanData['kar_permissions']),
          );
        } catch (e) {
          selectedPermissions = [];
        }
      }
    }

    bool isPremium = karyawanData?['kar_is_premium'] ?? false;
    bool isPasswordVisible = false;
    bool isPermissionsExpanded = true;

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => Dialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          backgroundColor: Colors.white,
          child: Container(
            width: MediaQuery.of(context).size.width * 0.9,
            constraints: BoxConstraints(
              maxHeight: MediaQuery.of(context).size.height * 0.85,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(16),
                      topRight: Radius.circular(16),
                    ),
                    border: Border(
                      bottom: BorderSide(color: Colors.grey[200]!),
                    ),
                  ),
                  child: Row(
                    children: [
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(Icons.arrow_back, color: Colors.black),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                      const SizedBox(width: 12),
                      Text(
                        isEdit ? 'Edit Data' : 'Tambah Pegawai',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.w600,
                          color: Colors.black,
                        ),
                      ),
                    ],
                  ),
                ),
                Flexible(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text(
                          'Nama Pegawai',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: Colors.black87,
                          ),
                        ),
                        const SizedBox(height: 8),
                        TextField(
                          controller: namaController,
                          decoration: InputDecoration(
                            hintText: 'Masukkan Nama Pegawai',
                            hintStyle: TextStyle(color: Colors.grey[400]),
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
                              borderSide: const BorderSide(
                                color: Color(0xFFFF6B6B),
                              ),
                            ),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          'No Handphone',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: Colors.black87,
                          ),
                        ),
                        const SizedBox(height: 8),
                        TextField(
                          controller: phoneController,
                          keyboardType: TextInputType.phone,
                          decoration: InputDecoration(
                            hintText: '0818-3573-2875',
                            hintStyle: TextStyle(color: Colors.grey[400]),
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
                              borderSide: const BorderSide(
                                color: Color(0xFFFF6B6B),
                              ),
                            ),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          'Email',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: Colors.black87,
                          ),
                        ),
                        const SizedBox(height: 8),
                        TextField(
                          controller: emailController,
                          keyboardType: TextInputType.emailAddress,
                          decoration: InputDecoration(
                            hintText: 'Masukkan Alamat Email',
                            hintStyle: TextStyle(color: Colors.grey[400]),
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
                              borderSide: const BorderSide(
                                color: Color(0xFFFF6B6B),
                              ),
                            ),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        Text(
                          isEdit
                              ? 'Password (kosongkan jika tidak diubah)'
                              : 'Password',
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: Colors.black87,
                          ),
                        ),
                        const SizedBox(height: 8),
                        TextField(
                          controller: passwordController,
                          obscureText: !isPasswordVisible,
                          decoration: InputDecoration(
                            hintText: 'Abijali123',
                            hintStyle: TextStyle(color: Colors.grey[400]),
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
                              borderSide: const BorderSide(
                                color: Color(0xFFFF6B6B),
                              ),
                            ),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16,
                              vertical: 12,
                            ),
                            suffixIcon: IconButton(
                              icon: Icon(
                                isPasswordVisible
                                    ? Icons.visibility_outlined
                                    : Icons.visibility_off_outlined,
                                color: Colors.grey[600],
                              ),
                              onPressed: () {
                                setDialogState(() {
                                  isPasswordVisible = !isPasswordVisible;
                                });
                              },
                            ),
                          ),
                        ),
                        const SizedBox(height: 20),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text(
                              'Lihat Akses Pegawai',
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: Colors.black87,
                              ),
                            ),
                            IconButton(
                              onPressed: () {
                                setDialogState(() {
                                  isPermissionsExpanded =
                                      !isPermissionsExpanded;
                                });
                              },
                              icon: Icon(
                                isPermissionsExpanded
                                    ? Icons.remove_circle_outline
                                    : Icons.add_circle_outline,
                                color: Colors.black87,
                                size: 20,
                              ),
                              padding: EdgeInsets.zero,
                              constraints: const BoxConstraints(),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        if (isPermissionsExpanded)
                          ...availablePermissions.map((permission) {
                            return _buildPermissionCheckbox(
                              title: permission,
                              value: selectedPermissions.contains(permission),
                              onChanged: (value) {
                                setDialogState(() {
                                  if (value ?? false) {
                                    selectedPermissions.add(permission);
                                  } else {
                                    selectedPermissions.remove(permission);
                                  }
                                });
                              },
                            );
                          }),
                      ],
                    ),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    border: Border(top: BorderSide(color: Colors.grey[200]!)),
                  ),
                  child: SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () {
                        if (namaController.text.isEmpty) {
                          _showSnackBar(
                            'Nama pegawai tidak boleh kosong',
                            Colors.red,
                          );
                          return;
                        }

                        if (phoneController.text.isEmpty) {
                          _showSnackBar(
                            'No handphone tidak boleh kosong',
                            Colors.red,
                          );
                          return;
                        }

                        if (emailController.text.isEmpty) {
                          _showSnackBar('Email tidak boleh kosong', Colors.red);
                          return;
                        }

                        if (!isEdit && passwordController.text.isEmpty) {
                          _showSnackBar(
                            'Password tidak boleh kosong',
                            Colors.red,
                          );
                          return;
                        }

                        if (!isEdit && passwordController.text.length < 6) {
                          _showSnackBar(
                            'Password minimal 6 karakter',
                            Colors.red,
                          );
                          return;
                        }

                        Navigator.pop(context);

                        _saveKaryawan(
                          karyawanId: karyawanData?['kar_id'],
                          nama: namaController.text,
                          phone: phoneController.text,
                          email: emailController.text,
                          password: passwordController.text,
                          isPremium: isPremium,
                          permissions: selectedPermissions,
                        );
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFFF6B6B),
                        padding: const EdgeInsets.symmetric(vertical: 14),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                        elevation: 0,
                      ),
                      child: const Text(
                        'SIMPAN DATA PEGAWAI',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          letterSpacing: 0.5,
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

  Widget _buildPermissionCheckbox({
    required String title,
    required bool value,
    required Function(bool?) onChanged,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Container(
            width: 20,
            height: 20,
            decoration: BoxDecoration(
              color: value ? const Color(0xFF4CAF50) : Colors.white,
              border: Border.all(
                color: value ? const Color(0xFF4CAF50) : Colors.grey[400]!,
                width: 2,
              ),
              borderRadius: BorderRadius.circular(4),
            ),
            child: InkWell(
              onTap: () => onChanged(!value),
              child: value
                  ? const Icon(Icons.check, size: 14, color: Colors.white)
                  : null,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              title,
              style: const TextStyle(fontSize: 13, color: Colors.black87),
            ),
          ),
        ],
      ),
    );
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
          style: TextStyle(
            color: Colors.black,
            fontSize: 16,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
      body: Column(
        children: [
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(color: Color(0xFFFF6B6B)),
                  )
                : karyawanList.isEmpty
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
                          'Belum ada data pegawai',
                          style: TextStyle(
                            fontSize: 16,
                            color: Colors.grey[600],
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Tambahkan pegawai pertama Anda',
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.grey[400],
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: karyawanList.length,
                    itemBuilder: (context, index) {
                      final karyawan = karyawanList[index];

                      List<String> permissions = [];
                      if (karyawan['kar_permissions'] != null) {
                        if (karyawan['kar_permissions'] is List) {
                          permissions = List<String>.from(
                            karyawan['kar_permissions'],
                          );
                        } else if (karyawan['kar_permissions'] is String) {
                          try {
                            permissions = List<String>.from(
                              json.decode(karyawan['kar_permissions']),
                            );
                          } catch (e) {
                            permissions = [];
                          }
                        }
                      }

                      final isExpanded =
                          expandedItems[karyawan['kar_id']] ?? false;

                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Container(
                                    width: 40,
                                    height: 40,
                                    decoration: BoxDecoration(
                                      color: Colors.grey[300],
                                      shape: BoxShape.circle,
                                    ),
                                    child: Icon(
                                      Icons.person_outline,
                                      color: Colors.grey[600],
                                      size: 24,
                                    ),
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          karyawan['kar_nama'] ??
                                              'Nama tidak tersedia',
                                          style: const TextStyle(
                                            fontSize: 15,
                                            fontWeight: FontWeight.w600,
                                            color: Colors.black,
                                          ),
                                        ),
                                        const SizedBox(height: 2),
                                        Text(
                                          'Ditambahkan pada: ${karyawan['kar_join_date']?.toString().split('T')[0] ?? ''}',
                                          style: TextStyle(
                                            fontSize: 11,
                                            color: Colors.grey[500],
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  OutlinedButton(
                                    onPressed: () => _showTambahKaryawanDialog(
                                      karyawanData: karyawan,
                                    ),
                                    style: OutlinedButton.styleFrom(
                                      side: const BorderSide(
                                        color: Color(0xFF4CAF50),
                                      ),
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 12,
                                        vertical: 6,
                                      ),
                                      minimumSize: Size.zero,
                                      tapTargetSize:
                                          MaterialTapTargetSize.shrinkWrap,
                                    ),
                                    child: const Text(
                                      'Edit',
                                      style: TextStyle(
                                        color: Color(0xFF4CAF50),
                                        fontSize: 12,
                                      ),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  OutlinedButton(
                                    onPressed: () => _showDeleteConfirmation(
                                      karyawan['kar_id'],
                                      karyawan['kar_nama'] ?? 'Karyawan',
                                    ),
                                    style: OutlinedButton.styleFrom(
                                      side: const BorderSide(color: Colors.red),
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 12,
                                        vertical: 6,
                                      ),
                                      minimumSize: Size.zero,
                                      tapTargetSize:
                                          MaterialTapTargetSize.shrinkWrap,
                                    ),
                                    child: const Text(
                                      'Hapus',
                                      style: TextStyle(
                                        color: Colors.red,
                                        fontSize: 12,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 12),
                              const Divider(height: 1),
                              const SizedBox(height: 12),
                              GestureDetector(
                                onTap: () {
                                  setState(() {
                                    expandedItems[karyawan['kar_id']] =
                                        !isExpanded;
                                  });
                                },
                                child: Row(
                                  mainAxisAlignment:
                                      MainAxisAlignment.spaceBetween,
                                  children: [
                                    const Text(
                                      'Lihat Akses Pegawai',
                                      style: TextStyle(
                                        fontSize: 13,
                                        fontWeight: FontWeight.w600,
                                        color: Colors.black87,
                                      ),
                                    ),
                                    Icon(
                                      isExpanded
                                          ? Icons.keyboard_arrow_up
                                          : Icons.keyboard_arrow_down,
                                      color: Colors.black87,
                                      size: 20,
                                    ),
                                  ],
                                ),
                              ),
                              if (isExpanded) ...[
                                const SizedBox(height: 12),
                                if (permissions.isEmpty)
                                  Padding(
                                    padding: const EdgeInsets.symmetric(
                                      vertical: 8,
                                    ),
                                    child: Text(
                                      'Tidak ada akses yang diberikan',
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: Colors.grey[500],
                                        fontStyle: FontStyle.italic,
                                      ),
                                    ),
                                  )
                                else
                                  ...permissions.map((permission) {
                                    return Padding(
                                      padding: const EdgeInsets.only(bottom: 8),
                                      child: Row(
                                        children: [
                                          Container(
                                            width: 16,
                                            height: 16,
                                            decoration: BoxDecoration(
                                              color: const Color(0xFF4CAF50),
                                              borderRadius:
                                                  BorderRadius.circular(3),
                                            ),
                                            child: const Icon(
                                              Icons.check,
                                              size: 12,
                                              color: Colors.white,
                                            ),
                                          ),
                                          const SizedBox(width: 8),
                                          Expanded(
                                            child: Text(
                                              permission,
                                              style: const TextStyle(
                                                fontSize: 12,
                                                color: Colors.black87,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    );
                                  }),
                              ],
                            ],
                          ),
                        ),
                      );
                    },
                  ),
          ),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
                  blurRadius: 10,
                  offset: const Offset(0, -5),
                ),
              ],
            ),
            child: SafeArea(
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => _showTambahKaryawanDialog(),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFF6B6B),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    elevation: 0,
                  ),
                  child: const Text(
                    'TAMBAHKAN PEGAWAI',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      letterSpacing: 0.5,
                    ),
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
