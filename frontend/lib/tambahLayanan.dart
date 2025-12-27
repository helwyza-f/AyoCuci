import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';
import 'package:flutter/services.dart';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';
import 'package:intl/intl.dart';
import 'package:http_parser/http_parser.dart';

const List<Map<String, dynamic>> customIcons = [
  {'icon': 'assets/icons/school-bag.png', 'label': 'Tas'},
  {'icon': 'assets/icons/shoes.png', 'label': 'Sepatu'},
  {'icon': 'assets/icons/cap.png', 'label': 'Topi'},
  {'icon': 'assets/icons/teddy-bear.png', 'label': 'Boneka'},
  {'icon': 'assets/icons/dress.png', 'label': 'Gaun'},
  {'icon': 'assets/icons/laundry.png', 'label': 'Laundry'},
  {'icon': 'assets/icons/wardrobe.png', 'label': 'Lemari'},
  {'icon': 'assets/icons/pants.png', 'label': 'Celana'},
  {'icon': 'assets/icons/wedding-dress.png', 'label': 'Gaun Pengantin'},
];

class CurrencyInputFormatter extends TextInputFormatter {
  final NumberFormat formatter = NumberFormat.decimalPattern('id_ID');

  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    if (newValue.text.isEmpty) {
      return newValue.copyWith(text: '');
    }

    String cleanText = newValue.text.replaceAll(RegExp(r'[^\d]'), '');
    if (cleanText.isEmpty) {
      return newValue.copyWith(text: '');
    }

    int value = int.parse(cleanText);
    String formattedText = formatter.format(value).replaceAll(',', '.');

    return TextEditingValue(
      text: formattedText,
      selection: TextSelection.collapsed(offset: formattedText.length),
    );
  }

  static int getCleanValue(TextEditingController controller) {
    String cleanText = controller.text.replaceAll(RegExp(r'[^\d]'), '');
    return int.tryParse(cleanText) ?? 0;
  }
}

class TambahLayananPage extends StatefulWidget {
  final String jenisLayanan;
  const TambahLayananPage({super.key, required this.jenisLayanan});

  @override
  State<TambahLayananPage> createState() => _TambahLayananPageState();
}

class _TambahLayananPageState extends State<TambahLayananPage> {
  final _namaLayananCtrl = TextEditingController();
  final _skorPrioritasCtrl = TextEditingController();
  final Map<String, bool> _selectedProses = {
    'Cuci': false,
    'Kering': false,
    'Setrika': false,
  };
  bool _isLoading = false;
  final List<Map<String, dynamic>> _jenisProdukList = [];
  final currencyFormatter = NumberFormat.decimalPattern('id_ID');

  @override
  void dispose() {
    _namaLayananCtrl.dispose();
    _skorPrioritasCtrl.dispose();
    super.dispose();
  }

  Future<Map<String, String>> _getAuthHeaders() async {
    final prefs = await SharedPreferences.getInstance();
    var token = prefs
        .getString('access_token')
        ?.replaceAll(RegExp(r'Bearer\s*|\n|\r'), '')
        .trim();
    return {'Authorization': 'Bearer ${token ?? ''}'};
  }

  void _showMsg(String msg, Color color) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  Future<void> _simpanLayanan() async {
    if (_namaLayananCtrl.text.trim().isEmpty) {
      return _showMsg('Nama layanan harus diisi', Colors.red);
    }
    final skor = int.tryParse(_skorPrioritasCtrl.text.trim());
    if (skor == null || skor < 0 || skor > 100) {
      return _showMsg('Skor prioritas harus angka 0-100', Colors.red);
    }
    final proses = _selectedProses.entries
        .where((e) => e.value)
        .map((e) => e.key)
        .toList();
    if (proses.isEmpty) {
      return _showMsg(
        'Pilih minimal 1 proses (Cuci/Kering/Setrika)',
        Colors.red,
      );
    }
    if (_jenisProdukList.isEmpty) {
      return _showMsg('Silahkan tambah minimal 1 Jenis Produk', Colors.red);
    }

    setState(() => _isLoading = true);

    try {
      final prefs = await SharedPreferences.getInstance();
      final headers = await _getAuthHeaders();
      final outletId = prefs.getInt('outletId');

      var request = http.MultipartRequest(
        'POST',
        Uri.parse('http://localhost:8080/v1/layanan/with-products'),
      );

      request.headers.addAll(headers);

      request.fields['ln_layanan'] = _namaLayananCtrl.text.trim();
      request.fields['ln_prioritas'] = skor.toString();
      request.fields['ln_cuci'] = proses.contains('Cuci') ? 'Ya' : 'Tidak';
      request.fields['ln_kering'] = proses.contains('Kering') ? 'Ya' : 'Tidak';
      request.fields['ln_setrika'] = proses.contains('Setrika')
          ? 'Ya'
          : 'Tidak';

      if (outletId != null) {
        request.fields['ln_outlet'] = outletId.toString();
      }

      List<Map<String, dynamic>> jenisProdukArray = [];

      for (int i = 0; i < _jenisProdukList.length; i++) {
        var product = _jenisProdukList[i];

        Map<String, dynamic> productData = {
          'jp_nama': product['nama'].toString(),
          'jp_satuan': product['satuan'].toString(),
          'jp_harga_per': product['harga'],
          'jp_lama_pengerjaan': product['lama_pengerjaan'],
          'jp_satuan_waktu': product['satuan_waktu'].toString(),
          'jp_keterangan': product['keterangan'] ?? '',
        };
        if (product['web_image_bytes'] != null) {
          final bytes = product['web_image_bytes'] as Uint8List;
          request.files.add(
            http.MultipartFile.fromBytes(
              'produk_$i', 
              bytes,
              filename:
                  'product_${DateTime.now().millisecondsSinceEpoch}_$i.jpg',
              contentType: MediaType('image', 'jpeg'),
            ),
          );
          productData['has_upload'] = true;
          productData['jp_icon_path'] = null;
        } else if (product['image_file_path'] != null &&
            product['image_file_path'].toString().isNotEmpty &&
            !kIsWeb) {
          final file = File(product['image_file_path']);
          if (await file.exists()) {
            final ext = file.path.split('.').last;
            request.files.add(
              await http.MultipartFile.fromPath(
                'produk_$i', 
                file.path,
                filename:
                    'product_${DateTime.now().millisecondsSinceEpoch}_$i.$ext',
                contentType: MediaType('image', ext == 'png' ? 'png' : 'jpeg'),
              ),
            );
            productData['has_upload'] = true;
            productData['jp_icon_path'] = null;
          }
        } else if (product['icon_path'] != null &&
            product['icon_path'].toString().isNotEmpty) {
          productData['jp_icon_path'] = product['icon_path'].toString();
          productData['has_upload'] = false;
        }

        jenisProdukArray.add(productData);
      }

      final jenisProdukJson = json.encode(jenisProdukArray);
      request.fields['jenis_produk'] = jenisProdukJson;

      print(' ========== RINGKASAN REQUEST ==========');
      print(' Endpoint: ${request.url}');
      print(' Files (${request.files.length}):');
      for (var file in request.files) {
        print(
          '  - Field: ${file.field}, Filename: ${file.filename}, Size: ${file.length} bytes',
        );
      }
      print('📤 Fields:');
      request.fields.forEach((key, value) {
        if (key == 'jenis_produk') {
          print(
            '  - $key: [JSON Array dengan ${jenisProdukArray.length} item]',
          );
        } else {
          print('  - $key: $value');
        }
      });
      print('================================================');

      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);

      print(' Response Status: ${response.statusCode}');
      print('Response Body: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        if (mounted) {
          _showMsg(' Layanan berhasil disimpan!', Colors.green);
          Navigator.pop(context, true);
        }
      } else {
        String errorMsg = 'Gagal menyimpan layanan';
        try {
          final responseData = json.decode(response.body);
          errorMsg = responseData['message'] ?? errorMsg;
        } catch (_) {
          errorMsg = 'Status ${response.statusCode}: ${response.body}';
        }
        _showMsg(' $errorMsg', Colors.red);
      }
    } catch (e, stackTrace) {
      print(' Exception: $e');
      print(' Stack trace: $stackTrace');
      _showMsg('Terjadi kesalahan: $e', Colors.red);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _tambahJenisProduk() async {
    final result = await showDialog<Map<String, dynamic>?>(
      context: context,
      builder: (_) => const ModalTambahProduk(),
    );

    if (result != null) {
      setState(() => _jenisProdukList.add(result));
      _showMsg('Produk ${result['nama']} berhasil ditambahkan', Colors.green);
    }
  }

  void _editJenisProduk(int index) async {
    final result = await showDialog<Map<String, dynamic>?>(
      context: context,
      builder: (_) => ModalTambahProduk(initialData: _jenisProdukList[index]),
    );

    if (result != null) {
      setState(() => _jenisProdukList[index] = result);
      _showMsg('Produk ${result['nama']} berhasil diubah', Colors.blue);
    }
  }

  void _hapusJenisProduk(int index) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text(
          'Hapus Produk',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        content: Text(
          'Apakah Anda yakin ingin menghapus ${_jenisProdukList[index]['nama']}?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Batal', style: TextStyle(color: Colors.grey)),
          ),
          ElevatedButton(
            onPressed: () {
              setState(() => _jenisProdukList.removeAt(index));
              Navigator.pop(context);
              _showMsg('Produk berhasil dihapus', Colors.orange);
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(8),
              ),
            ),
            child: const Text('Hapus', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  Widget _buildTextField(
    String label,
    TextEditingController ctrl, {
    bool isNumber = false,
    TextInputFormatter? formatter,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 8),
        TextField(
          controller: ctrl,
          keyboardType: isNumber ? TextInputType.number : TextInputType.text,
          inputFormatters: formatter != null
              ? [formatter]
              : (isNumber ? [FilteringTextInputFormatter.digitsOnly] : null),
          decoration: InputDecoration(
            hintText: isNumber ? 'Range 0 - 100' : 'Kiloan, Satuan, Horden',
            hintStyle: TextStyle(color: Colors.grey[400], fontSize: 13),
            filled: true,
            fillColor: Colors.white,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey[300]!),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide(color: Colors.grey[300]!),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: Color(0xFFFF7043), width: 2),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 14,
            ),
          ),
        ),
        const SizedBox(height: 20),
      ],
    );
  }

  Widget _buildCheckbox(String label) {
    return Expanded(
      child: InkWell(
        onTap: () => setState(
          () => _selectedProses[label] = !(_selectedProses[label] ?? false),
        ),
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: _selectedProses[label] == true
                ? const Color(0xFFFF7043).withOpacity(0.1)
                : Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: _selectedProses[label] == true
                  ? const Color(0xFFFF7043)
                  : Colors.grey[300]!,
              width: 1.5,
            ),
          ),
          child: Row(
            children: [
              Icon(
                _selectedProses[label] == true
                    ? Icons.check_circle
                    : Icons.circle_outlined,
                size: 20,
                color: _selectedProses[label] == true
                    ? const Color(0xFFFF7043)
                    : Colors.grey,
              ),
              const SizedBox(width: 8),
              Text(
                label,
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: _selectedProses[label] == true
                      ? FontWeight.w600
                      : FontWeight.normal,
                  color: _selectedProses[label] == true
                      ? const Color(0xFFFF7043)
                      : Colors.black87,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildProdukCard(Map<String, dynamic> product, int index) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey[200]!),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 70,
            height: 70,
            margin: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: const Color(0xFFFF7043).withOpacity(0.05),
              borderRadius: BorderRadius.circular(10),
            ),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(10),
              child: _buildProductImage(product),
            ),
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 10),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    product['nama'],
                    style: const TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    'Rp ${currencyFormatter.format(product['harga'])} / ${product['satuan']}',
                    style: TextStyle(
                      fontSize: 13,
                      color: Colors.grey[700],
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  Text(
                    '${product['lama_pengerjaan']} ${product['satuan_waktu']}',
                    style: TextStyle(fontSize: 12, color: Colors.grey[500]),
                  ),
                ],
              ),
            ),
          ),
          IconButton(
            icon: Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.blue.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(Icons.edit, color: Colors.blue, size: 18),
            ),
            onPressed: () => _editJenisProduk(index),
          ),
          IconButton(
            icon: Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.red.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(Icons.delete, color: Colors.red, size: 18),
            ),
            onPressed: () => _hapusJenisProduk(index),
          ),
          const SizedBox(width: 8),
        ],
      ),
    );
  }

  Widget _buildProductImage(Map<String, dynamic> product) {
    if (product['icon_path'] != null &&
        product['icon_path'].toString().isNotEmpty) {
      return Image.asset(
        product['icon_path'],
        fit: BoxFit.contain,
        errorBuilder: (context, error, stackTrace) =>
            const Icon(Icons.image, size: 30, color: Colors.grey),
      );
    } else if (product['web_image_bytes'] != null) {
      return Image.memory(
        product['web_image_bytes'] as Uint8List,
        fit: BoxFit.cover,
      );
    } else if (product['image_file_path'] != null &&
        product['image_file_path'].toString().isNotEmpty) {
      if (kIsWeb) {
        return const Icon(Icons.image, size: 30, color: Colors.grey);
      } else {
        return Image.file(
          File(product['image_file_path']),
          fit: BoxFit.cover,
          errorBuilder: (context, error, stackTrace) =>
              const Icon(Icons.image, size: 30, color: Colors.grey),
        );
      }
    }
    return const Icon(Icons.image, size: 30, color: Colors.grey);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[200],
      appBar: AppBar(
        backgroundColor: Colors.transparent,
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
      body: Stack(
        children: [
          Positioned(
            top: 20,
            left: 0,
            right: 0,
            bottom: 0,
            child: Column(
              children: [
                Container(
                  height: 30,
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.only(
                      topLeft: Radius.circular(1000),
                      topRight: Radius.circular(1000),
                    ),
                  ),
                ),
                Expanded(
                  child: Container(
                    color: Colors.white,
                    child: SingleChildScrollView(
                      child: Padding(
                        padding: const EdgeInsets.all(24),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text(
                              'Tambah Layanan Baru',
                              style: TextStyle(
                                fontSize: 24,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              'Isi informasi layanan dan produk yang tersedia',
                              style: TextStyle(
                                fontSize: 14,
                                color: Colors.grey[600],
                              ),
                            ),
                            const SizedBox(height: 24),
                            _buildTextField('Nama Layanan', _namaLayananCtrl),
                            _buildTextField(
                              'Skor Prioritas',
                              _skorPrioritasCtrl,
                              isNumber: true,
                            ),
                            const Text(
                              'Proses',
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                            const SizedBox(height: 12),
                            Row(
                              children: [
                                'Cuci',
                                'Kering',
                                'Setrika',
                              ].map((e) => _buildCheckbox(e)).toList(),
                            ),
                            const SizedBox(height: 28),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                const Text(
                                  'Jenis Produk',
                                  style: TextStyle(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                  ),
                                ),
                                ElevatedButton.icon(
                                  onPressed: _tambahJenisProduk,
                                  icon: const Icon(Icons.add, size: 18),
                                  label: const Text('Tambah Produk'),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: const Color(0xFFFF7043),
                                    foregroundColor: Colors.white,
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 16,
                                      vertical: 12,
                                    ),
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(10),
                                    ),
                                    elevation: 0,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 16),
                            _jenisProdukList.isEmpty
                                ? Container(
                                    width: double.infinity,
                                    padding: const EdgeInsets.all(24),
                                    decoration: BoxDecoration(
                                      color: Colors.grey[50],
                                      borderRadius: BorderRadius.circular(12),
                                      border: Border.all(
                                        color: Colors.grey[200]!,
                                      ),
                                    ),
                                    child: Column(
                                      children: [
                                        Icon(
                                          Icons.inventory_2_outlined,
                                          size: 48,
                                          color: Colors.grey[400],
                                        ),
                                        const SizedBox(height: 12),
                                        Text(
                                          'Belum ada produk',
                                          style: TextStyle(
                                            color: Colors.grey[700],
                                            fontSize: 16,
                                            fontWeight: FontWeight.w600,
                                          ),
                                        ),
                                        const SizedBox(height: 4),
                                        Text(
                                          'Tambahkan minimal 1 jenis produk',
                                          style: TextStyle(
                                            color: Colors.grey[500],
                                            fontSize: 13,
                                          ),
                                        ),
                                      ],
                                    ),
                                  )
                                : Column(
                                    children: List.generate(
                                      _jenisProdukList.length,
                                      (index) => _buildProdukCard(
                                        _jenisProdukList[index],
                                        index,
                                      ),
                                    ),
                                  ),
                            const SizedBox(height: 32),
                            SizedBox(
                              width: double.infinity,
                              child: ElevatedButton(
                                onPressed: _isLoading ? null : _simpanLayanan,
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: const Color(0xFFFF7043),
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 16,
                                  ),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                  elevation: 0,
                                ),
                                child: _isLoading
                                    ? const SizedBox(
                                        height: 20,
                                        width: 20,
                                        child: CircularProgressIndicator(
                                          color: Colors.white,
                                          strokeWidth: 2,
                                        ),
                                      )
                                    : const Text(
                                        'SIMPAN LAYANAN',
                                        style: TextStyle(
                                          color: Colors.white,
                                          fontSize: 15,
                                          fontWeight: FontWeight.bold,
                                          letterSpacing: 0.5,
                                        ),
                                      ),
                              ),
                            ),
                            const SizedBox(height: 24),
                          ],
                        ),
                      ),
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
}

class ModalTambahProduk extends StatefulWidget {
  final Map<String, dynamic>? initialData;
  const ModalTambahProduk({super.key, this.initialData});

  @override
  State<ModalTambahProduk> createState() => _ModalTambahProdukState();
}

class _ModalTambahProdukState extends State<ModalTambahProduk> {
  final _namaProdukCtrl = TextEditingController();
  final _hargaCtrl = TextEditingController();
  final _lamaPengerjaanCtrl = TextEditingController();
  final _keteranganCtrl = TextEditingController();
  final _hargaFormatter = CurrencyInputFormatter();

  String? _selectedIconPath;
  String _selectedSatuan = 'kg';
  String _selectedSatuanWaktu = 'Hari';
  String? _imageFilePath;
  Uint8List? _webImageBytes;
  String? _imageUrl;
  String? _imageFileName;
  int? _productId;

  final List<String> _satuanOptions = [
    'kg',
    'pcs',
    'm2',
    'bh',
    'unit',
    'pasang',
  ];
  final List<String> _satuanWaktuOptions = ['Menit', 'Jam', 'Hari'];
  final ImagePicker _picker = ImagePicker();

  @override
  void initState() {
    super.initState();
    if (widget.initialData != null) {
      _loadInitialData(widget.initialData!);
    }
  }

  void _loadInitialData(Map<String, dynamic> data) {
    _productId = data['id'];
    _namaProdukCtrl.text = data['nama'] ?? '';
    final hargaText = data['harga']?.toString() ?? '';

    _hargaCtrl.text = hargaText.isNotEmpty
        ? CurrencyInputFormatter()
              .formatEditUpdate(
                TextEditingValue.empty,
                TextEditingValue(text: hargaText),
              )
              .text
        : '';

    _lamaPengerjaanCtrl.text = data['lama_pengerjaan']?.toString() ?? '';
    _keteranganCtrl.text = data['keterangan'] ?? '';

    _selectedSatuan = data['satuan'] ?? 'kg';
    _selectedSatuanWaktu = data['satuan_waktu'] ?? 'Hari';

    _selectedIconPath = data['icon_path'];
    _imageFilePath = data['image_file_path'];
    _webImageBytes = data['web_image_bytes'];
    _imageUrl = data['image_url'];
    _imageFileName = data['image_file_name'];
  }

  Future<void> _pickImage() async {
    final XFile? pickedFile = await _picker.pickImage(
      source: ImageSource.gallery,
    );

    if (pickedFile != null) {
      setState(() {
        _selectedIconPath = null;
        _imageUrl = null;

        if (kIsWeb) {
          pickedFile.readAsBytes().then((bytes) {
            setState(() {
              _webImageBytes = bytes;
              _imageFilePath = null;
              _imageFileName = pickedFile.name;
            });
          });
        } else {
          _imageFilePath = pickedFile.path;
          _webImageBytes = null;
          _imageFileName = null;
        }
      });
    }
  }

  void _removeImage() {
    setState(() {
      _imageFilePath = null;
      _webImageBytes = null;
      _imageUrl = null;
      _imageFileName = null;
    });
  }

  void _selectIcon(String path) {
    setState(() {
      _selectedIconPath = path;
      _removeImage();
    });
  }

  Widget _buildTextField(
    String label,
    TextEditingController ctrl, {
    String? hint,
    TextInputType keyboardType = TextInputType.text,
    List<TextInputFormatter>? formatters,
    bool isRequired = true,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '$label ${isRequired ? '*' : ''}',
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          TextField(
            controller: ctrl,
            keyboardType: keyboardType,
            inputFormatters: formatters,
            decoration: InputDecoration(
              hintText: hint,
              hintStyle: TextStyle(color: Colors.grey[400], fontSize: 13),
              filled: true,
              fillColor: Colors.grey[50],
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
                borderSide: BorderSide(color: Colors.grey[300]!),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
                borderSide: BorderSide(color: Colors.grey[300]!),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(10),
                borderSide: const BorderSide(
                  color: Color(0xFFFF7043),
                  width: 2,
                ),
              ),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 14,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDropdown(
    String label,
    String value,
    List<String> options,
    ValueChanged<String?> onChanged,
  ) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '$label *',
            style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            decoration: BoxDecoration(
              color: Colors.grey[50],
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: Colors.grey[300]!),
            ),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                value: value,
                isExpanded: true,
                icon: const Icon(
                  Icons.arrow_drop_down,
                  color: Color(0xFFFF7043),
                ),
                style: const TextStyle(color: Colors.black87, fontSize: 15),
                onChanged: onChanged,
                items: options.map<DropdownMenuItem<String>>((String val) {
                  return DropdownMenuItem<String>(value: val, child: Text(val));
                }).toList(),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProductImageSelection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text(
          'Ikon / Gambar Produk',
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 10),

        Center(
          child: Container(
            width: 100,
            height: 100,
            decoration: BoxDecoration(
              color: Colors.grey[100],
              borderRadius: BorderRadius.circular(15),
              border: Border.all(color: Colors.grey[300]!),
            ),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(15),
              child: _getCurrentImageWidget(100, 100),
            ),
          ),
        ),
        const SizedBox(height: 10),

        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            ElevatedButton.icon(
              onPressed: _pickImage,
              icon: const Icon(Icons.camera_alt, size: 18),
              label: const Text('Pilih Gambar'),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blueAccent,
                foregroundColor: Colors.white,
              ),
            ),
            if (_selectedIconPath != null ||
                _imageFilePath != null ||
                _webImageBytes != null ||
                _imageUrl != null)
              Padding(
                padding: const EdgeInsets.only(left: 8.0),
                child: ElevatedButton.icon(
                  onPressed: _removeImage,
                  icon: const Icon(Icons.close, size: 18),
                  label: const Text('Hapus'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.red,
                    foregroundColor: Colors.white,
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 20),

        const Text(
          'Atau Pilih Ikon Default',
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
        ),
        const SizedBox(height: 10),
        GridView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 5,
            crossAxisSpacing: 10,
            mainAxisSpacing: 10,
          ),
          itemCount: customIcons.length,
          itemBuilder: (context, index) {
            final iconData = customIcons[index];
            final isSelected = _selectedIconPath == iconData['icon'];
            return InkWell(
              onTap: () => _selectIcon(iconData['icon']),
              borderRadius: BorderRadius.circular(10),
              child: Container(
                decoration: BoxDecoration(
                  color: isSelected
                      ? const Color(0xFFFF7043).withOpacity(0.2)
                      : Colors.white,
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(
                    color: isSelected
                        ? const Color(0xFFFF7043)
                        : Colors.grey[300]!,
                    width: isSelected ? 2 : 1,
                  ),
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Image.asset(iconData['icon'], width: 30, height: 30),
                    const SizedBox(height: 4),
                    Text(
                      iconData['label'],
                      style: TextStyle(
                        fontSize: 10,
                        color: isSelected
                            ? const Color(0xFFFF7043)
                            : Colors.black87,
                      ),
                    ),
                  ],
                ),
              ),
            );
          },
        ),
        const SizedBox(height: 20),
      ],
    );
  }

  Widget _getCurrentImageWidget(double width, double height) {
    if (_selectedIconPath != null) {
      return Image.asset(
        _selectedIconPath!,
        fit: BoxFit.contain,
        width: width * 0.7,
        height: height * 0.7,
        errorBuilder: (context, error, stackTrace) =>
            const Icon(Icons.image_not_supported, color: Colors.grey),
      );
    } else if (_webImageBytes != null) {
      return Image.memory(
        _webImageBytes!,
        fit: BoxFit.cover,
        width: width,
        height: height,
      );
    } else if (_imageFilePath != null && !kIsWeb) {
      return Image.file(
        File(_imageFilePath!),
        fit: BoxFit.cover,
        width: width,
        height: height,
        errorBuilder: (context, error, stackTrace) =>
            const Icon(Icons.image_not_supported, color: Colors.grey),
      );
    } else if (_imageUrl != null) {
      return Image.network(
        'http://localhost:8080/$_imageUrl',
        fit: BoxFit.cover,
        width: width,
        height: height,
        errorBuilder: (context, error, stackTrace) =>
            const Icon(Icons.image_not_supported, color: Colors.grey),
      );
    }
    return Icon(Icons.inventory_2_outlined, size: 40, color: Colors.grey[400]);
  }

  void _saveProduct() {
    final trimmedNama = _namaProdukCtrl.text.trim();
    if (trimmedNama.isEmpty) {
      return _showMsg('Nama produk harus diisi', Colors.red);
    }
    final hargaBersih = CurrencyInputFormatter.getCleanValue(_hargaCtrl);
    if (hargaBersih <= 0) {
      return _showMsg('Harga produk harus lebih dari 0', Colors.red);
    }
    final lamaPengerjaan = int.tryParse(_lamaPengerjaanCtrl.text.trim());
    if (lamaPengerjaan == null || lamaPengerjaan <= 0) {
      return _showMsg('Lama pengerjaan harus diisi dan > 0', Colors.red);
    }
    if (_selectedIconPath == null &&
        _imageFilePath == null &&
        _webImageBytes == null &&
        _imageUrl == null) {
      return _showMsg('Pilih ikon atau upload gambar produk', Colors.red);
    }

    final Map<String, dynamic> result = {
      'id': _productId,
      'nama': trimmedNama,
      'satuan': _selectedSatuan,
      'harga': hargaBersih,
      'lama_pengerjaan': lamaPengerjaan,
      'satuan_waktu': _selectedSatuanWaktu,
      'keterangan': _keteranganCtrl.text.trim(),

      'icon_path': _selectedIconPath,
      'image_file_path': _imageFilePath,
      'web_image_bytes': _webImageBytes,
      'image_url': _imageUrl,
      'image_file_name': _imageFileName,
      'should_delete': false,
    };

    Navigator.pop(context, result);
  }

  void _showMsg(String msg, Color color) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.initialData == null
                    ? 'Tambah Jenis Produk'
                    : 'Edit Jenis Produk',
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Color(0xFFFF7043),
                ),
              ),
              const Divider(height: 30),

              _buildTextField(
                'Nama Produk',
                _namaProdukCtrl,
                hint: 'Contoh: Pakaian, Selimut, Sprei',
              ),

              _buildTextField(
                'Harga Per ${(_selectedSatuan).toUpperCase()}',
                _hargaCtrl,
                hint: 'Contoh: 8000',
                keyboardType: TextInputType.number,
                formatters: [_hargaFormatter],
              ),

              _buildDropdown('Satuan', _selectedSatuan, _satuanOptions, (
                String? newValue,
              ) {
                if (newValue != null) {
                  setState(() => _selectedSatuan = newValue);
                }
              }),

              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: _buildTextField(
                      'Lama Pengerjaan',
                      _lamaPengerjaanCtrl,
                      hint: 'Contoh: 1, 2, 30',
                      keyboardType: TextInputType.number,
                      formatters: [FilteringTextInputFormatter.digitsOnly],
                    ),
                  ),
                  const SizedBox(width: 15),
                  Expanded(
                    child: _buildDropdown(
                      'Satuan Waktu',
                      _selectedSatuanWaktu,
                      _satuanWaktuOptions,
                      (String? newValue) {
                        if (newValue != null) {
                          setState(() => _selectedSatuanWaktu = newValue);
                        }
                      },
                    ),
                  ),
                ],
              ),

              _buildTextField(
                'Keterangan (Opsional)',
                _keteranganCtrl,
                hint: 'Catatan tambahan',
                isRequired: false,
              ),

              _buildProductImageSelection(),

              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _saveProduct,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFFFF7043),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 14),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                  child: Text(
                    widget.initialData == null
                        ? 'SIMPAN PRODUK'
                        : 'UBAH PRODUK',
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
              ),

              SizedBox(
                width: double.infinity,
                child: TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text(
                    'Batal',
                    style: TextStyle(color: Colors.grey),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
