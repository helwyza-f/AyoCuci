import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class KategoriPengeluaranPage extends StatefulWidget {
  const KategoriPengeluaranPage({super.key});

  @override
  State<KategoriPengeluaranPage> createState() =>
      _KategoriPengeluaranPageState();
}

class _KategoriPengeluaranPageState extends State<KategoriPengeluaranPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  List<Map<String, dynamic>> kategoriList = [];
  bool _isLoading = false;
  String? _accessToken;
  int? _outletId;
  final TextEditingController _searchController = TextEditingController();
  List<Map<String, dynamic>> _filteredKategoriList = [];

  @override
  void initState() {
    super.initState();
    _loadData();
    _searchController.addListener(_filterKategori);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _filterKategori() {
    final query = _searchController.text.toLowerCase();
    setState(() {
      if (query.isEmpty) {
        _filteredKategoriList = kategoriList;
      } else {
        _filteredKategoriList = kategoriList
            .where(
              (kategori) =>
                  (kategori['ktg_nama'] ?? '').toLowerCase().contains(query),
            )
            .toList();
      }
    });
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
        await _fetchKategoriData();
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

  Future<void> _fetchKategoriData() async {
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
      final url = '$_baseUrl/kategori-pengeluaran';

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
          kategoriList = rawList;
          _filteredKategoriList = rawList;
          _isLoading = false;
        });
      } else if (response.statusCode == 401) {
        setState(() => _isLoading = false);
        _showSnackBar('Token tidak valid. Silakan login kembali', Colors.red);
      } else {
        setState(() => _isLoading = false);
        _showSnackBar('Gagal memuat data kategori', Colors.red);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _saveKategori({
    int? kategoriId,
    required String namaKategori,
  }) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    if (_outletId == null) {
      _showSnackBar(
        'Outlet ID tidak ditemukan. Silakan login kembali',
        Colors.red,
      );
      return;
    }

    try {
      final url = kategoriId == null
          ? '$_baseUrl/kategori-pengeluaran'
          : '$_baseUrl/kategori-pengeluaran/$kategoriId';

      final body = {'ktg_nama': namaKategori};

      final response = kategoriId == null
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
          kategoriId == null
              ? 'Kategori berhasil ditambahkan'
              : 'Kategori berhasil diupdate',
          Colors.green,
        );
        await _fetchKategoriData();
      } else {
        final errorData = json.decode(response.body);
        final errorMessage = errorData['message'] ?? 'Gagal menyimpan kategori';
        _showSnackBar(errorMessage, Colors.red);
      }
    } catch (e) {
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _deleteKategori(int kategoriId) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      return;
    }

    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/kategori-pengeluaran/$kategoriId'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      if (response.statusCode == 200) {
        _showSnackBar('Kategori berhasil dihapus', Colors.green);
        await _fetchKategoriData();
      } else {
        _showSnackBar('Gagal menghapus kategori', Colors.red);
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

  void _showDeleteConfirmation(int kategoriId, String namaKategori) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Hapus'),
        content: Text(
          'Apakah Anda yakin ingin menghapus kategori "$namaKategori"?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteKategori(kategoriId);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  void _showTambahKategoriDialog({Map<String, dynamic>? kategoriData}) {
    final isEdit = kategoriData != null;
    final namaController = TextEditingController(
      text: kategoriData?['ktg_nama'] ?? '',
    );

    showDialog(
      context: context,
      builder: (context) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
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
                      isEdit ? 'Edit Kategori' : 'Tambah Kategori',
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

                // Nama Kategori
                const Text(
                  'Nama Kategori',
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
                ),
                const SizedBox(height: 8),
                TextField(
                  controller: namaController,
                  decoration: InputDecoration(
                    hintText: 'Contoh: Gaji, Bonus, Listrik',
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
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: () {
                      if (namaController.text.isEmpty) {
                        _showSnackBar(
                          'Nama kategori tidak boleh kosong',
                          Colors.red,
                        );
                        return;
                      }

                      Navigator.pop(context);

                      _saveKategori(
                        kategoriId: kategoriData?['ktg_id'],
                        namaKategori: namaController.text,
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
      body: Column(
        children: [
          Container(
            color: Colors.white,
            padding: const EdgeInsets.all(16),
            child: TextField(
              controller: _searchController,
              decoration: InputDecoration(
                hintText: 'Cari Pengeluaran',
                hintStyle: TextStyle(color: Colors.grey[400]),
                prefixIcon: Icon(Icons.search, color: Colors.grey[400]),
                filled: true,
                fillColor: Colors.grey[100],
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide.none,
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide.none,
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
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(
                      valueColor: AlwaysStoppedAnimation<Color>(
                        Color(0xFFFF6B6B),
                      ),
                    ),
                  )
                : _filteredKategoriList.isEmpty
                ? Center(
                    child: Padding(
                      padding: const EdgeInsets.all(32),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            Icons.category_outlined,
                            size: 64,
                            color: Colors.grey[300],
                          ),
                          const SizedBox(height: 16),
                          Text(
                            _searchController.text.isEmpty
                                ? 'Belum ada kategori'
                                : 'Kategori tidak ditemukan',
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
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _filteredKategoriList.length,
                    itemBuilder: (context, index) {
                      return _buildKategoriCard(_filteredKategoriList[index]);
                    },
                  ),
          ),
          Container(
            color: Colors.white,
            padding: const EdgeInsets.all(16),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => _showTambahKategoriDialog(),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFFF6B6B),
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  elevation: 0,
                ),
                child: const Text(
                  'TAMBAH KATEGORI',
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
    );
  }

  Widget _buildKategoriCard(Map<String, dynamic> kategori) {
    final nama = kategori['ktg_nama'] ?? 'Kategori';
    final kategoriId = kategori['ktg_id'];

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              nama,
              style: const TextStyle(
                fontSize: 15,
                fontWeight: FontWeight.w500,
                color: Colors.black87,
              ),
            ),
          ),
          const SizedBox(width: 12),
          OutlinedButton(
            onPressed: () => _showTambahKategoriDialog(kategoriData: kategori),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
              side: const BorderSide(color: Color(0xFF4CAF50)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(6),
              ),
              minimumSize: const Size(50, 32),
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
            onPressed: () => _showDeleteConfirmation(kategoriId, nama),
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
              side: const BorderSide(color: Color(0xFFEF5350)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(6),
              ),
              minimumSize: const Size(50, 32),
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
