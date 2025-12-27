import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class ParfumPage extends StatefulWidget {
  const ParfumPage({super.key});

  @override
  State<ParfumPage> createState() => _ParfumPageState();
}

class _ParfumPageState extends State<ParfumPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  List<Map<String, dynamic>> parfumList = [];
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
        await _fetchParfumData();
      } else {
        _showSnackBar('Outlet ID tidak ditemukan. Silakan login kembali', Colors.red);
      }
    } else {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
    }
  }

  Future<void> _fetchParfumData() async {
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
      final url = '$_baseUrl/parfum';

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
          parfumList = rawList;
          _isLoading = false;
        });
      } else if (response.statusCode == 401) {
        setState(() => _isLoading = false);
        _showSnackBar('Token tidak valid. Silakan login kembali', Colors.red);
      } else {
        setState(() => _isLoading = false);
        _showSnackBar('Gagal memuat data parfum', Colors.red);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _saveParfum({
    int? parfumId,
    required String namaParfum,
    required String keterangan,
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
      final url = parfumId == null
          ? '$_baseUrl/parfum'
          : '$_baseUrl/parfum/$parfumId';

      final body = {
        'prf_nama': namaParfum,
        'prf_keterangan': keterangan,
      };

      print('Request URL: $url');
      print('Request body: ${json.encode(body)}');

      final response = parfumId == null
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
          parfumId == null
              ? 'Parfum berhasil ditambahkan'
              : 'Parfum berhasil diupdate',
          Colors.green,
        );
        await _fetchParfumData();
      } else {
        final errorData = json.decode(response.body);
        final errorMessage = errorData['message'] ?? 'Gagal menyimpan parfum';
        _showSnackBar(errorMessage, Colors.red);
      }
    } catch (e) {
      print('Error: $e');
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _deleteParfum(int parfumId) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/parfum/$parfumId'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      if (response.statusCode == 200) {
        _showSnackBar('Parfum berhasil dihapus', Colors.green);
        await _fetchParfumData();
      } else {
        _showSnackBar('Gagal menghapus parfum', Colors.red);
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

  void _showDeleteConfirmation(int parfumId, String namaParfum) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Hapus'),
        content: Text(
          'Apakah Anda yakin ingin menghapus parfum "$namaParfum"?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteParfum(parfumId);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  void _showTambahParfumDialog({Map<String, dynamic>? parfumData}) {
    final isEdit = parfumData != null;
    final namaController = TextEditingController(
      text: parfumData?['prf_nama'] ?? '',
    );
    final keteranganController = TextEditingController(
      text: parfumData?['prf_keterangan'] ?? '',
    );

    showDialog(
      context: context,
      builder: (context) => Dialog(
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
                      isEdit ? 'Edit Parfum' : 'Tambah Parfum',
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

                // Nama Parfum
                const Text(
                  'Nama Parfum',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),
                TextField(
                  controller: namaController,
                  decoration: InputDecoration(
                    hintText: 'Downy',
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
                    hintText: 'Dengan berat 12 Kg',
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
                const SizedBox(height: 24),

                // Tombol Simpan
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: () {
                      if (namaController.text.isEmpty) {
                        _showSnackBar(
                          'Nama parfum tidak boleh kosong',
                          Colors.red,
                        );
                        return;
                      }

                      Navigator.pop(context);

                      _saveParfum(
                        parfumId: parfumData?['prf_id'],
                        namaParfum: namaController.text,
                        keterangan: keteranganController.text,
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
            // Header dengan badge total
            Container(
              margin: const EdgeInsets.all(16),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'List Parfum',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w600,
                      color: Colors.black87,
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFF6B6B),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Text(
                      'Total: ${parfumList.length}',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Tambah Parfum Button
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: () => _showTambahParfumDialog(),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFF6B6B),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    elevation: 0,
                  ),
                  child: const Text(
                    'TAMBAH PARFUM',
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
            const SizedBox(height: 24),

            // Parfum List
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: _isLoading
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
                  : parfumList.isEmpty
                      ? Center(
                          child: Padding(
                            padding: const EdgeInsets.all(32),
                            child: Column(
                              children: [
                                Icon(
                                  Icons.water_drop_outlined,
                                  size: 64,
                                  color: Colors.grey[300],
                                ),
                                const SizedBox(height: 16),
                                Text(
                                  'Belum ada parfum',
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
                          children: parfumList
                              .map((parfum) => _buildParfumCard(parfum))
                              .toList(),
                        ),
            ),
            const SizedBox(height: 16),
          ],
        ),
      ),
    );
  }

  Widget _buildParfumCard(Map<String, dynamic> parfum) {
    final nama = parfum['prf_nama'] ?? 'Parfum';
    final keterangan = parfum['prf_keterangan'] ?? '';
    final parfumId = parfum['prf_id'];

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.1),
            spreadRadius: 1,
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  nama,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: Colors.black87,
                  ),
                ),
                if (keterangan.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text(
                    keterangan,
                    style: TextStyle(
                      fontSize: 13,
                      color: Colors.grey[600],
                    ),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(width: 12),
          OutlinedButton(
            onPressed: () => _showTambahParfumDialog(parfumData: parfum),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 8,
              ),
              side: const BorderSide(color: Color(0xFF4CAF50)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              minimumSize: const Size(60, 36),
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
            onPressed: () => _showDeleteConfirmation(parfumId, nama),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 8,
              ),
              side: const BorderSide(color: Color(0xFFEF5350)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              minimumSize: const Size(60, 36),
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
    );
  }
}