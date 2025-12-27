import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'editLayanan.dart';
import 'tambahLayanan.dart';

class LayananPage extends StatefulWidget {
  final String jenisLayanan;

  const LayananPage({super.key, required this.jenisLayanan});

  @override
  State<LayananPage> createState() => _LayananPageState();
}

class _LayananPageState extends State<LayananPage> {
  final String _baseUrl = 'http://localhost:8080/v1';
  List<Map<String, dynamic>> _layananList = [];
  bool _isLoading = false;
  String? _accessToken;

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

    setState(() {
      _accessToken = token;
    });

    if (_accessToken != null && _accessToken!.isNotEmpty) {
      await _fetchLayananData();
    } else {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      _redirectToLogin();
    }
  }

  Future<void> _fetchLayananData() async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak valid. Silakan login kembali', Colors.red);
      _redirectToLogin();
      return;
    }

    setState(() => _isLoading = true);

    try {
      final url = '$_baseUrl/layanan';

      print('=== FETCH REQUEST ===');
      print('URL: $url');
      print('Token length: ${_accessToken!.length}');
      print('====================');

      final response = await http.get(
        Uri.parse(url),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      print('=== FETCH RESPONSE ===');
      print('Status Code: ${response.statusCode}');
      print('Response Body: ${response.body}');
      print('=====================');

      if (!mounted) return;

      if (response.statusCode == 200) {
        // Clear image cache untuk refresh gambar
        imageCache.clear();
        imageCache.clearLiveImages();

        final data = json.decode(response.body);
        final rawList = List<Map<String, dynamic>>.from(data['data'] ?? []);

        print('=== DATA PARSED ===');
        print('Total items: ${rawList.length}');
        if (rawList.isNotEmpty) {
          print('First item: ${json.encode(rawList[0])}');
        }
        print('==================');

        setState(() {
          _layananList = rawList;
          _isLoading = false;
        });
        print('✓ Successfully loaded ${_layananList.length} items');
      } else if (response.statusCode == 401) {
        setState(() => _isLoading = false);
        await _clearTokenAndRedirect();
      } else {
        setState(() => _isLoading = false);
        _showSnackBar('Gagal memuat data layanan', Colors.red);
      }
    } catch (e) {
      setState(() => _isLoading = false);
      print('❌ Exception during fetch: $e');
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  Future<void> _clearTokenAndRedirect() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('access_token');
    await prefs.remove('refresh_token');
    _redirectToLogin();
  }

  void _redirectToLogin() {
    if (mounted) {
      Future.delayed(const Duration(milliseconds: 500), () {
        if (mounted) {
          Navigator.of(context).popUntil((route) => route.isFirst);
        }
      });
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

  Future<void> _deleteLayanan(int layananId) async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      await _clearTokenAndRedirect();
      return;
    }

    try {
      final response = await http.delete(
        Uri.parse('$_baseUrl/layanan/$layananId'),
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $_accessToken',
        },
      );

      if (response.statusCode == 200) {
        _showSnackBar('Layanan berhasil dihapus', Colors.green);
        await _fetchLayananData();
      } else if (response.statusCode == 401) {
        await _clearTokenAndRedirect();
      } else {
        _showSnackBar('Gagal menghapus layanan', Colors.red);
      }
    } catch (e) {
      _showSnackBar('Terjadi kesalahan: ${e.toString()}', Colors.red);
    }
  }

  void _showDeleteConfirmation(int layananId, String namaLayanan) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Konfirmasi Hapus'),
        content: Text(
          'Apakah Anda yakin ingin menghapus layanan "$namaLayanan"?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              _deleteLayanan(layananId);
            },
            style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  Future<void> _navigateToTambahLayanan() async {
    if (_accessToken == null || _accessToken!.isEmpty) {
      _showSnackBar('Token tidak ditemukan. Silakan login kembali', Colors.red);
      _redirectToLogin();
      return;
    }

    final result = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            TambahLayananPage(jenisLayanan: widget.jenisLayanan),
      ),
    );

    if (result == true) {
      await _fetchLayananData();
    }
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
          // Tambah Layanan Button
          Container(
            color: Colors.white,
            padding: const EdgeInsets.all(16),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _navigateToTambahLayanan,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFFF6B6B),
                  padding: const EdgeInsets.symmetric(vertical: 14),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  elevation: 0,
                ),
                child: const Text(
                  'TAMBAH LAYANAN',
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

          // Content Area
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(
                      valueColor: AlwaysStoppedAnimation<Color>(
                        Color(0xFFFF6B6B),
                      ),
                    ),
                  )
                : _layananList.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.inbox_outlined,
                          size: 64,
                          color: Colors.grey[300],
                        ),
                        const SizedBox(height: 16),
                        Text(
                          'Belum ada layanan',
                          style: TextStyle(
                            color: Colors.grey[600],
                            fontSize: 16,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Tambahkan layanan untuk memulai',
                          style: TextStyle(
                            color: Colors.grey[400],
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _layananList.length,
                    itemBuilder: (context, index) {
                      final layanan = _layananList[index];
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 12),
                        child: _buildLayananCard(layanan),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildLayananCard(Map<String, dynamic> layanan) {
    final namaLayanan = layanan['ln_layanan'] ?? 'Layanan';
    final layananId = layanan['id'];
    var jenisProdukList = layanan['jenis_produk'] as List<dynamic>? ?? [];

    final score = layanan['ln_prioritas'] ?? 0;

    final cuci = layanan['ln_cuci']?.toString() ?? '';
    final kering = layanan['ln_kering']?.toString() ?? '';
    final setrika = layanan['ln_setrika']?.toString() ?? '';

    List<String> prosesList = [];
    if (cuci.isNotEmpty && cuci.toLowerCase() != 'tidak')
      prosesList.add('Cuci');
    if (kering.isNotEmpty && kering.toLowerCase() != 'tidak')
      prosesList.add('Kering');
    if (setrika.isNotEmpty && setrika.toLowerCase() != 'tidak')
      prosesList.add('Setrika');

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.05),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header Section
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: const Color(0xFFF8F9FA),
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(12),
                topRight: Radius.circular(12),
              ),
            ),
            child: Row(
              children: [
                Text(
                  namaLayanan.toString(),
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: Colors.black87,
                  ),
                ),
                const Spacer(),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.grey[300]!),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Text(
                        'Scores: ',
                        style: TextStyle(fontSize: 12, color: Colors.black54),
                      ),
                      Text(
                        score.toString(),
                        style: const TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                          color: Colors.black87,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 8,
                    vertical: 4,
                  ),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFEBEE),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(
                        Icons.circle,
                        size: 8,
                        color: Color(0xFFFF6B6B),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        prosesList.isEmpty
                            ? 'Tidak Ada'
                            : prosesList.join(' + '),
                        style: const TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: Color(0xFFFF6B6B),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),

          // Product List Section
          if (jenisProdukList.isEmpty)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
              decoration: BoxDecoration(
                border: Border(
                  bottom: BorderSide(color: Colors.grey[200]!, width: 1),
                ),
              ),
              child: Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF0F0F0),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.local_laundry_service_outlined,
                      color: Color(0xFF666666),
                      size: 22,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          namaLayanan,
                          style: const TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: Colors.black87,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Belum ada detail produk',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[500],
                            fontStyle: FontStyle.italic,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            )
          else
            ...jenisProdukList.map((produk) {
              final namaProduk = produk['jp_nama'] ?? 'Produk';
              final harga = produk['jp_harga_per'] ?? 0;
              final satuan = produk['jp_satuan'] ?? 'Kg';
              final durasi = produk['jp_lama_pengerjaan'] ?? 0;
              final satuanWaktu = produk['jp_satuan_waktu'] ?? 'Hari';

              // ✅ PERBAIKAN: Ambil jp_image_url dan jp_icon_path dari API
              final imageUrl = produk['jp_image_url']?.toString() ?? '';
              final iconPath = produk['jp_icon_path']?.toString() ?? '';
              final lastUpdate = produk['jp_lastupdate']?.toString() ?? '';

              print('=== IMAGE URL DEBUG ===');
              print('Product: $namaProduk');
              print('jp_image_url: $imageUrl');
              print('jp_icon_path: $iconPath');
              print('=====================');

              return Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 12,
                ),
                decoration: BoxDecoration(
                  border: Border(
                    bottom: BorderSide(color: Colors.grey[200]!, width: 1),
                  ),
                ),
                child: Row(
                  children: [
                    // Product Icon
                    Container(
                      key: ValueKey('${produk['id']}_$lastUpdate'),
                      width: 40,
                      height: 40,
                      decoration: BoxDecoration(
                        color: const Color(0xFFF0F0F0),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: _buildProductImage(imageUrl, iconPath),
                      ),
                    ),
                    const SizedBox(width: 12),

                    // Product Info
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            namaProduk,
                            style: const TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.w600,
                              color: Colors.black87,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Rp ${harga.toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (Match m) => '${m[1]}.')} / $satuan • $durasi $satuanWaktu',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey[600],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            }).toList(),
        
          // Action Buttons
          Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (context) => EditLayananPage(
                            layananId: layananId,
                            layananData: layanan,
                          ),
                        ),
                      ).then((result) {
                        if (result == true) {
                          _fetchLayananData();
                        }
                      });
                    },
                    icon: const Icon(
                      Icons.edit_outlined,
                      size: 16,
                      color: Color(0xFF4CAF50),
                    ),
                    label: const Text(
                      'Edit Layanan',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Color(0xFF4CAF50),
                      ),
                    ),
                    style: OutlinedButton.styleFrom(
                      side: const BorderSide(
                        color: Color(0xFF4CAF50),
                        width: 1.5,
                      ),
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () {
                      _showDeleteConfirmation(layananId, namaLayanan);
                    },
                    icon: const Icon(Icons.delete_outline, size: 16),
                    label: const Text(
                      'Hapus Layanan',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFFFF6B6B),
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 10),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      elevation: 0,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// ✅ METHOD BARU: Build Product Image dengan logika yang benar
  Widget _buildProductImage(String imageUrl, String iconPath) {
    final baseServer = _baseUrl.replaceAll('/v1', ''); // http://localhost:8080

    // Prioritas 1: Jika ada jp_image_url (uploaded image)
    if (imageUrl.isNotEmpty) {
      final fullImageUrl = imageUrl.startsWith('http')
          ? imageUrl
          : '$baseServer/$imageUrl';

      print('✅ Using jp_image_url: $fullImageUrl');

      return Image.network(
        '$fullImageUrl?v=${DateTime.now().millisecondsSinceEpoch}',
        width: 40,
        height: 40,
        fit: BoxFit.cover,
        cacheWidth: 120,
        cacheHeight: 120,
        errorBuilder: (context, error, stackTrace) {
          print('❌ Error loading image: $fullImageUrl');
          print('Error: $error');
          return const Icon(
            Icons.local_laundry_service_outlined,
            color: Color(0xFF666666),
            size: 22,
          );
        },
        loadingBuilder: (context, child, loadingProgress) {
          if (loadingProgress == null) return child;
          return Center(
            child: CircularProgressIndicator(
              value: loadingProgress.expectedTotalBytes != null
                  ? loadingProgress.cumulativeBytesLoaded /
                      loadingProgress.expectedTotalBytes!
                  : null,
              strokeWidth: 2,
              valueColor: const AlwaysStoppedAnimation<Color>(
                Color(0xFFFF6B6B),
              ),
            ),
          );
        },
      );
    }

    // Prioritas 2: Jika ada jp_icon_path (asset icon)
    if (iconPath.isNotEmpty && iconPath.startsWith('assets/')) {
      print('✅ Using asset icon: $iconPath');
      return Image.asset(
        iconPath,
        width: 40,
        height: 40,
        fit: BoxFit.contain,
        errorBuilder: (context, error, stackTrace) {
          print('❌ Error loading asset: $iconPath');
          return const Icon(
            Icons.local_laundry_service_outlined,
            color: Color(0xFF666666),
            size: 22,
          );
        },
      );
    }

    // Prioritas 3: Default icon
    print('⚠️ No image or icon found, using default');
    return const Icon(
      Icons.local_laundry_service_outlined,
      color: Color(0xFF666666),
      size: 22,
    );
  }

  @override
  void dispose() {
    super.dispose();
  }
}