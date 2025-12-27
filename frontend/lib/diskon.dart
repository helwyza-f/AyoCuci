import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class DiskonPage extends StatefulWidget {
  const DiskonPage({super.key});

  @override
  State<DiskonPage> createState() => _DiskonPageState();
}

class _DiskonPageState extends State<DiskonPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  bool fiturDiskonAktif = true;
  List<Map<String, dynamic>> diskonList = [];
  bool _isLoading = false;
  String? _accessToken;
  int? _outletId;

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
        await _fetchDiskonData();
      } else {
        _showSnackBar('Outlet ID tidak ditemukan. Silakan login kembali', Colors.red);
      }
    } else {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
    }
  }

  Future<void> _fetchDiskonData() async {
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
      final url = '$_baseUrl/diskon';

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
          diskonList = rawList;
          _isLoading = false;
        });
      } else if (response.statusCode == 401) {
        setState(() => _isLoading = false);
        _showSnackBar('Token tidak valid. Silakan login kembali', Colors.red);
      } else {
        setState(() => _isLoading = false);
        _showSnackBar('Gagal memuat data diskon', Colors.red);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _saveDiskon({
    int? diskonId,
    required String namaDiskon,
    required String jenisDiskon,
    required double nilaiDiskon,
    required String keterangan,
    required bool status,
  }) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    if (_outletId == null) {
      _showSnackBar('Outlet ID tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    try {
      final url = diskonId == null
          ? '$_baseUrl/diskon'
          : '$_baseUrl/diskon/$diskonId';

      // Sesuaikan dengan struct DiskonInput di backend (dengan prefix dis_)
      final body = {
        'dis_diskon': namaDiskon,
        'dis_jenis': jenisDiskon,
        'dis_nilai_diskon': nilaiDiskon,
        'dis_keterangan': keterangan,
        'dis_status': status ? 'Aktif' : 'Tidak Aktif',
      };

      print('Request URL: $url');
      print('Request body: ${json.encode(body)}');

      final response = diskonId == null
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

      print('Response status: ${response.statusCode}');
      print('Response body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        _showSnackBar(
          diskonId == null
              ? 'Diskon berhasil ditambahkan'
              : 'Diskon berhasil diupdate',
          Colors.green,
        );
        await _fetchDiskonData();
      } else {
        final errorData = json.decode(response.body);
        final errorMessage = errorData['message'] ?? 'Gagal menyimpan diskon';
        _showSnackBar(errorMessage, Colors.red);
      }
    } catch (e) {
      print('Error: $e');
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _deleteDiskon(int diskonId) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/diskon/$diskonId'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      if (response.statusCode == 200) {
        _showSnackBar('Diskon berhasil dihapus', Colors.green);
        await _fetchDiskonData();
      } else {
        _showSnackBar('Gagal menghapus diskon', Colors.red);
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

  void _showDeleteConfirmation(int diskonId, String namaDiskon) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Hapus'),
        content: Text(
          'Apakah Anda yakin ingin menghapus diskon "$namaDiskon"?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteDiskon(diskonId);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  void _showTambahDiskonDialog({Map<String, dynamic>? diskonData}) {
    final isEdit = diskonData != null;
    final namaController = TextEditingController(
      text: diskonData?['dis_diskon'] ?? '',
    );
    final nilaiController = TextEditingController(
      text: diskonData?['dis_nilai_diskon']?.toString() ?? '',
    );
    final keteranganController = TextEditingController(
      text: diskonData?['dis_keterangan'] ?? '',
    );

    String jenisDiskon = diskonData?['dis_jenis'] ?? 'Nominal';
    bool status = (diskonData?['dis_status'] ?? 'Aktif') == 'Aktif';

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
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
            ),
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        isEdit ? 'Edit Diskon' : 'Tambah Diskon',
                        style: const TextStyle(
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      IconButton(
                        onPressed: () => Navigator.pop(context),
                        icon: const Icon(Icons.close),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),

                  // Nama Diskon
                  const Text(
                    'Nama Diskon',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: namaController,
                    decoration: InputDecoration(
                      hintText: 'Diskon Grand Opening',
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
                        borderSide: const BorderSide(color: Color(0xFFFF6B6B)),
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 12,
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Jenis Diskon
                  const Text(
                    'Jenis Diskon',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      Expanded(
                        child: RadioListTile<String>(
                          title: const Text(
                            'Nominal',
                            style: TextStyle(fontSize: 14),
                          ),
                          value: 'Nominal',
                          groupValue: jenisDiskon,
                          onChanged: (value) {
                            setDialogState(() {
                              jenisDiskon = value!;
                            });
                          },
                          activeColor: const Color(0xFFFF6B6B),
                          contentPadding: EdgeInsets.zero,
                          dense: true,
                        ),
                      ),
                      Expanded(
                        child: RadioListTile<String>(
                          title: const Text(
                            'Persen',
                            style: TextStyle(fontSize: 14),
                          ),
                          value: 'Persen',
                          groupValue: jenisDiskon,
                          onChanged: (value) {
                            setDialogState(() {
                              jenisDiskon = value!;
                            });
                          },
                          activeColor: const Color(0xFFFF6B6B),
                          contentPadding: EdgeInsets.zero,
                          dense: true,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Nilai Diskon
                  const Text(
                    'Nilai Diskon',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: nilaiController,
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(
                      hintText: jenisDiskon == 'Nominal' ? '10000' : '10',
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
                        borderSide: const BorderSide(color: Color(0xFFFF6B6B)),
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 12,
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Keterangan
                  const Text(
                    'Keterangan',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: keteranganController,
                    maxLines: 2,
                    decoration: InputDecoration(
                      hintText: 'Diskon dari 25 Nov - 01 Des',
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
                        borderSide: const BorderSide(color: Color(0xFFFF6B6B)),
                      ),
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 12,
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Status
                  const Text(
                    'Status',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      color: status
                          ? const Color(0xFFFFEBEE)
                          : Colors.grey[200],
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          status ? 'Aktif' : 'Tidak Aktif',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: status
                                ? const Color(0xFFFF6B6B)
                                : Colors.grey[600],
                          ),
                        ),
                        Switch(
                          value: status,
                          onChanged: (value) {
                            setDialogState(() {
                              status = value;
                            });
                          },
                          activeColor: Colors.white,
                          activeTrackColor: const Color(0xFFFF6B6B),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Tombol Simpan
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton(
                      onPressed: () {
                        if (namaController.text.isEmpty) {
                          _showSnackBar(
                            'Nama diskon tidak boleh kosong',
                            Colors.red,
                          );
                          return;
                        }

                        final nilai = double.tryParse(nilaiController.text);
                        if (nilai == null) {
                          _showSnackBar(
                            'Nilai diskon harus berupa angka',
                            Colors.red,
                          );
                          return;
                        }

                        Navigator.pop(context);

                        _saveDiskon(
                          diskonId: diskonData?['dis_id'],
                          namaDiskon: namaController.text,
                          jenisDiskon: jenisDiskon,
                          nilaiDiskon: nilai,
                          keterangan: keteranganController.text,
                          status: status,
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
                        'SIMPAN',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          letterSpacing: 0.5,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[100],
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
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Fitur Diskon Section
            Container(
              margin: const EdgeInsets.all(16),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFFFE5E5),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Fitur Diskon',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: Colors.black87,
                            ),
                          ),
                          SizedBox(height: 4),
                          Text(
                            'Fitur Diskon Akan Muncul dalam menu transaksi',
                            style: TextStyle(
                              fontSize: 11,
                              color: Colors.black54,
                            ),
                          ),
                        ],
                      ),
                      CircleAvatar(
                        backgroundColor: Colors.white,
                        radius: 14,
                        child: Icon(
                          Icons.help_outline,
                          size: 18,
                          color: Colors.grey[600],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Aktif',
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                          color: Colors.black87,
                        ),
                      ),
                      Switch(
                        value: fiturDiskonAktif,
                        onChanged: (value) {
                          setState(() {
                            fiturDiskonAktif = value;
                          });
                        },
                        activeColor: Colors.white,
                        activeTrackColor: const Color(0xFFEF5350),
                      ),
                    ],
                  ),
                ],
              ),
            ),

            // Tambah Diskon Button
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => _showTambahDiskonDialog(),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFF6B6B),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    elevation: 0,
                  ),
                  child: const Text(
                    'TAMBAH DISKON',
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
            const SizedBox(height: 16),

            // Diskon Saat Ini Section
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Diskon Saat Ini',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: Colors.black87,
                    ),
                  ),
                  const SizedBox(height: 4),
                  const Text(
                    'Diskon yang tersedia saat ini',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey,
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Diskon List
                  _isLoading
                      ? const Center(
                          child: Padding(
                            padding: EdgeInsets.all(32),
                            child: CircularProgressIndicator(
                              valueColor: AlwaysStoppedAnimation<Color>(
                                Color(0xFFFF6B6B),
                              ),
                            ),
                          ),
                        )
                      : diskonList.isEmpty
                          ? Center(
                              child: Padding(
                                padding: const EdgeInsets.all(32),
                                child: Column(
                                  children: [
                                    Icon(
                                      Icons.discount_outlined,
                                      size: 64,
                                      color: Colors.grey[300],
                                    ),
                                    const SizedBox(height: 16),
                                    Text(
                                      'Belum ada diskon',
                                      style: TextStyle(
                                        color: Colors.grey[600],
                                        fontSize: 16,
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            )
                          : Column(
                              children: diskonList
                                  .map((diskon) => _buildDiskonCard(diskon))
                                  .toList(),
                            ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildDiskonCard(Map<String, dynamic> diskon) {
    final nama = diskon['dis_diskon'] ?? 'Diskon';
    final jenis = diskon['dis_jenis'] ?? 'Nominal';
    final nilai = diskon['dis_nilai_diskon'] ?? 0;
    final keterangan = diskon['dis_keterangan'] ?? '';
    final aktif = (diskon['dis_status'] ?? 'Aktif') == 'Aktif';
    final diskonId = diskon['dis_id'];

    String potongan;
    if (jenis == 'Persen') {
      potongan = '$nilai% / Transaksi';
    } else {
      potongan = 'Rp ${nilai.toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (Match m) => '${m[1]}.')} / Transaksi';
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: aktif ? const Color(0xFF4CAF50) : Colors.grey,
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 8),
              Text(
                aktif ? 'Aktif' : 'Tidak Aktif',
                style: const TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w500,
                  color: Colors.black87,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Text(
            nama,
            style: const TextStyle(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: Colors.black87,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            potongan,
            style: const TextStyle(
              fontSize: 12,
              color: Colors.grey,
            ),
          ),
          if (keterangan.isNotEmpty) ...[
            const SizedBox(height: 4),
            Text(
              keterangan,
              style: TextStyle(
                fontSize: 11,
                color: Colors.grey[600],
                fontStyle: FontStyle.italic,
              ),
            ),
          ],
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              OutlinedButton(
                onPressed: () => _showTambahDiskonDialog(diskonData: diskon),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 8,
                  ),
                  side: const BorderSide(color: Color(0xFF4CAF50)),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                child: const Text(
                  'Edit',
                  style: TextStyle(
                    color: Color(0xFF4CAF50),
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                onPressed: () => _showDeleteConfirmation(diskonId, nama),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 20,
                    vertical: 8,
                  ),
                  side: const BorderSide(color: Color(0xFFEF5350)),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
                child: const Text(
                  'Hapus',
                  style: TextStyle(
                    color: Color(0xFFEF5350),
                    fontSize: 13,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}