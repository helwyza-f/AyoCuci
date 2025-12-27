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

const List<Map<String, String>> customIcons = [
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

const List<String> satuanOptions = ['kg', 'pcs', 'm2', 'bh', 'unit', 'pasang'];
const List<String> satuanWaktuOptions = ['Menit', 'Jam', 'Hari', 'Minggu'];

class CurrencyInputFormatter extends TextInputFormatter {
  final NumberFormat formatter = NumberFormat.decimalPattern('id_ID');

  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    if (newValue.text.isEmpty) return newValue.copyWith(text: '');
    String cleanText = newValue.text.replaceAll(RegExp(r'[^\d]'), '');
    if (cleanText.isEmpty) return newValue.copyWith(text: '');
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

class ModalTambahProduk extends StatefulWidget {
  final Map<String, dynamic>? initialData;
  final String baseUrl;

  const ModalTambahProduk({super.key, this.initialData, required this.baseUrl});

  @override
  State<ModalTambahProduk> createState() => _ModalTambahProdukState();
}

class _ModalTambahProdukState extends State<ModalTambahProduk> {
  final _formKey = GlobalKey<FormState>();
  final _namaCtrl = TextEditingController();
  final _hargaCtrl = TextEditingController();
  final _lamaPengerjaanCtrl = TextEditingController();
  final _keteranganCtrl = TextEditingController();

  String _selectedSatuan = 'kg';
  String _selectedSatuanWaktu = 'Hari';
  String? _selectedAssetPath;
  String? _imageFilePath;
  Uint8List? _webImageBytes;
  String? _initialImageUrl;

  @override
  void initState() {
    super.initState();
    if (widget.initialData != null) {
      _namaCtrl.text = widget.initialData!['nama'] ?? '';
      _hargaCtrl.text = CurrencyInputFormatter().formatter
          .format(widget.initialData!['harga'] ?? 0)
          .replaceAll(',', '.');
      _lamaPengerjaanCtrl.text =
          widget.initialData!['lama_pengerjaan']?.toString() ?? '0';
      _keteranganCtrl.text = widget.initialData!['keterangan'] ?? '';
      _selectedSatuan = widget.initialData!['satuan'] ?? 'kg';
      _selectedSatuanWaktu = widget.initialData!['satuan_waktu'] ?? 'Hari';
      _initialImageUrl = widget.initialData!['image_url'];
      _selectedAssetPath = widget.initialData!['icon_path'];
      _imageFilePath = widget.initialData!['image_file_path'];
      _webImageBytes = widget.initialData!['web_image_bytes'];
    }
  }

  @override
  void dispose() {
    _namaCtrl.dispose();
    _hargaCtrl.dispose();
    _lamaPengerjaanCtrl.dispose();
    _keteranganCtrl.dispose();
    super.dispose();
  }

  Future<void> _pickImage(ImageSource source) async {
    final picker = ImagePicker();
    final pickedFile = await picker.pickImage(source: source, imageQuality: 50);

    if (pickedFile != null) {
      if (kIsWeb) {
        final bytes = await pickedFile.readAsBytes();
        setState(() {
          _selectedAssetPath = null;
          _initialImageUrl = null;
          _imageFilePath = null;
          _webImageBytes = bytes;
        });
      } else {
        setState(() {
          _selectedAssetPath = null;
          _initialImageUrl = null;
          _webImageBytes = null;
          _imageFilePath = pickedFile.path;
        });
      }
      if (mounted) Navigator.pop(context);
    }
  }

  void _showImageSourceDialog() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => Container(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Pilih Gambar Produk',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 20),
            _buildImageOption(
              icon: Icons.photo_library_outlined,
              title: 'Pilih dari Galeri',
              onTap: () => _pickImage(ImageSource.gallery),
            ),
            _buildImageOption(
              icon: Icons.category_outlined,
              title: 'Pilih Ikon',
              onTap: () {
                Navigator.pop(context);
                _showIconSelectionDialog();
              },
            ),
            if (_selectedAssetPath != null ||
                _imageFilePath != null ||
                _webImageBytes != null ||
                _initialImageUrl != null)
              _buildImageOption(
                icon: Icons.delete_outline,
                title: 'Hapus Gambar',
                color: Colors.red,
                onTap: () {
                  setState(() {
                    _selectedAssetPath = null;
                    _imageFilePath = null;
                    _webImageBytes = null;
                    _initialImageUrl = null;
                  });
                  Navigator.pop(context);
                },
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildImageOption({
    required IconData icon,
    required String title,
    required VoidCallback onTap,
    Color? color,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
        decoration: BoxDecoration(
          border: Border(bottom: BorderSide(color: Colors.grey[200]!)),
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: (color ?? const Color(0xFFFF6B6B)).withOpacity(0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Icon(
                icon,
                color: color ?? const Color(0xFFFF6B6B),
                size: 24,
              ),
            ),
            const SizedBox(width: 16),
            Text(
              title,
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
                color: color ?? Colors.black87,
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _showIconSelectionDialog() {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.white,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => Container(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text(
              'Pilih Ikon',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 20),
            SizedBox(
              height: 300,
              child: GridView.builder(
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 4,
                  crossAxisSpacing: 12,
                  mainAxisSpacing: 12,
                  childAspectRatio: 0.85,
                ),
                itemCount: customIcons.length,
                itemBuilder: (context, index) {
                  final iconData = customIcons[index];
                  final isSelected = _selectedAssetPath == iconData['icon'];
                  return InkWell(
                    onTap: () {
                      setState(() {
                        _selectedAssetPath = iconData['icon'];
                        _initialImageUrl = null;
                        _imageFilePath = null;
                        _webImageBytes = null;
                      });
                      Navigator.pop(context);
                    },
                    child: Container(
                      decoration: BoxDecoration(
                        color: isSelected
                            ? const Color(0xFFFF6B6B).withOpacity(0.1)
                            : Colors.grey[50],
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: isSelected
                              ? const Color(0xFFFF6B6B)
                              : Colors.grey[300]!,
                          width: isSelected ? 2.5 : 1,
                        ),
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Image.asset(
                            iconData['icon']!,
                            width: 36,
                            height: 36,
                            fit: BoxFit.contain,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            iconData['label']!,
                            style: TextStyle(
                              fontSize: 11,
                              color: isSelected
                                  ? const Color(0xFFFF6B6B)
                                  : Colors.black87,
                            ),
                            textAlign: TextAlign.center,
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _submitProduct() {
    if (!_formKey.currentState!.validate()) return;
    final priceClean = CurrencyInputFormatter.getCleanValue(_hargaCtrl);
    Navigator.pop(context, {
      'id': widget.initialData?['id'],
      'nama': _namaCtrl.text.trim(),
      'satuan': _selectedSatuan,
      'harga': priceClean,
      'lama_pengerjaan': int.tryParse(_lamaPengerjaanCtrl.text) ?? 0,
      'satuan_waktu': _selectedSatuanWaktu,
      'keterangan': _keteranganCtrl.text.trim(),
      'icon_path': _selectedAssetPath,
      'image_url': _initialImageUrl,
      'image_file_path': _imageFilePath,
      'web_image_bytes': _webImageBytes,
    });
  }

  Widget _buildImagePreview() {
    final baseServer = widget.baseUrl.replaceAll('/v1', '');
    Widget imageWidget;

    if (_selectedAssetPath != null) {
      imageWidget = Image.asset(_selectedAssetPath!, fit: BoxFit.contain);
    } else if (_webImageBytes != null) {
      imageWidget = Image.memory(_webImageBytes!, fit: BoxFit.cover);
    } else if (_imageFilePath != null) {
      imageWidget = kIsWeb
          ? const Icon(Icons.image, size: 50, color: Colors.white70)
          : Image.file(File(_imageFilePath!), fit: BoxFit.cover);
    } else if (_initialImageUrl != null && _initialImageUrl!.isNotEmpty) {
      final fullImageUrl = _initialImageUrl!.startsWith('http')
          ? _initialImageUrl!
          : '$baseServer/$_initialImageUrl';
      imageWidget = Image.network(
        fullImageUrl,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) =>
            const Icon(Icons.broken_image, size: 50, color: Colors.white70),
      );
    } else {
      imageWidget = const Icon(
        Icons.add_photo_alternate_outlined,
        size: 50,
        color: Colors.white70,
      );
    }

    return GestureDetector(
      onTap: _showImageSourceDialog,
      child: Container(
        width: 120,
        height: 120,
        decoration: BoxDecoration(
          color: const Color(0xFFFF6B6B).withOpacity(0.15),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: const Color(0xFFFF6B6B).withOpacity(0.3),
            width: 2,
          ),
        ),
        child: Stack(
          children: [
            Positioned.fill(
              child: ClipRRect(
                borderRadius: BorderRadius.circular(14),
                child: imageWidget,
              ),
            ),
            Positioned(
              bottom: 8,
              right: 8,
              child: Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: Colors.black.withOpacity(0.6),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(Icons.edit, color: Colors.white, size: 18),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      child: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      widget.initialData == null ? 'Tambah Produk' : 'Edit Produk',
                      style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
                const SizedBox(height: 20),
                Center(child: _buildImagePreview()),
                const SizedBox(height: 24),
                _buildTextField('Nama Produk', _namaCtrl, hint: 'Contoh: Pakaian Dewasa'),
                _buildTextField('Harga Per Satuan', _hargaCtrl, isCurrency: true, hint: 'Masukkan harga'),
                _buildDropdown('Satuan', _selectedSatuan, satuanOptions, (v) => setState(() => _selectedSatuan = v!)),
                Row(
                  children: [
                    Expanded(
                      flex: 2,
                      child: _buildTextField('Lama Pengerjaan', _lamaPengerjaanCtrl, isNumber: true, hint: 'Cth: 2'),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      flex: 3,
                      child: _buildDropdown('Satuan Waktu', _selectedSatuanWaktu, satuanWaktuOptions, (v) => setState(() => _selectedSatuanWaktu = v!)),
                    ),
                  ],
                ),
                _buildTextField('Keterangan (Opsional)', _keteranganCtrl, maxLines: 3, hint: 'Tambahkan keterangan'),
                const SizedBox(height: 16),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _submitProduct,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFFFF6B6B),
                      padding: const EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                    ),
                    child: Text(
                      widget.initialData == null ? 'TAMBAH PRODUK' : 'SIMPAN',
                      style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
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

  Widget _buildTextField(String label, TextEditingController ctrl, {bool isNumber = false, bool isCurrency = false, String? hint, int maxLines = 1}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          TextFormField(
            controller: ctrl,
            keyboardType: isNumber || isCurrency ? TextInputType.number : TextInputType.text,
            inputFormatters: isCurrency ? [CurrencyInputFormatter()] : (isNumber ? [FilteringTextInputFormatter.digitsOnly] : null),
            maxLines: maxLines,
            decoration: InputDecoration(
              hintText: hint,
              prefixText: isCurrency ? 'Rp ' : null,
              filled: true,
              fillColor: Colors.grey[50],
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey[300]!)),
              enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey[300]!)),
              focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: Color(0xFFFF6B6B), width: 2)),
              contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            ),
            validator: (v) => (v == null || v.trim().isEmpty) && !label.contains('Opsional') ? 'Wajib diisi' : null,
          ),
        ],
      ),
    );
  }

  Widget _buildDropdown(String label, String value, List<String> items, ValueChanged<String?> onChanged) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            value: value,
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.grey[50],
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey[300]!)),
              enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: Colors.grey[300]!)),
              focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: const BorderSide(color: Color(0xFFFF6B6B), width: 2)),
              contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            ),
            items: items.map((i) => DropdownMenuItem(value: i, child: Text(i))).toList(),
            onChanged: onChanged,
          ),
        ],
      ),
    );
  }
}

class EditLayananPage extends StatefulWidget {
  final int layananId;
  final Map<String, dynamic> layananData;

  const EditLayananPage({super.key, required this.layananId, required this.layananData});

  @override
  State<EditLayananPage> createState() => _EditLayananPageState();
}

class _EditLayananPageState extends State<EditLayananPage> {
  final _formKey = GlobalKey<FormState>();
  final String _baseUrl = 'http://localhost:8080/v1';
  final _namaLayananCtrl = TextEditingController();
  final _skorPrioritasCtrl = TextEditingController();
  final Map<String, bool> _selectedProses = {'Cuci': false, 'Kering': false, 'Setrika': false};
  List<Map<String, dynamic>> _produkList = [];
  bool _isLoading = false;
  String? _accessToken;
  final currencyFormatter = NumberFormat.decimalPattern('id_ID');

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  void _loadInitialData() async {
    final prefs = await SharedPreferences.getInstance();
    String? token = prefs.getString('access_token');
    if (token != null) {
      token = token.trim().replaceFirst('Bearer ', '').replaceAll(RegExp(r'[\n\r]'), '');
    }
    setState(() {
      _accessToken = token;
      _namaLayananCtrl.text = widget.layananData['ln_layanan'] ?? '';
      _skorPrioritasCtrl.text = widget.layananData['ln_prioritas']?.toString() ?? '0';
      _selectedProses['Cuci'] = (widget.layananData['ln_cuci']?.toString().toLowerCase() ?? 'tidak') == 'ya';
      _selectedProses['Kering'] = (widget.layananData['ln_kering']?.toString().toLowerCase() ?? 'tidak') == 'ya';
      _selectedProses['Setrika'] = (widget.layananData['ln_setrika']?.toString().toLowerCase() ?? 'tidak') == 'ya';
      var jenisProduk = widget.layananData['jenis_produk'] as List<dynamic>? ?? [];
      _produkList = jenisProduk.map((p) => _normalizeProduk(p)).toList();
    });
  }

  Map<String, dynamic> _normalizeProduk(dynamic p) {
    String satuan = p['jp_satuan']?.toString().toLowerCase() ?? 'kg';
    if (!satuanOptions.contains(satuan)) satuan = 'kg';
    String satuanWaktu = p['jp_satuan_waktu']?.toString() ?? 'Hari';
    if (satuanWaktu.isNotEmpty) {
      satuanWaktu = satuanWaktu[0].toUpperCase() + satuanWaktu.substring(1).toLowerCase();
    }
    if (!satuanWaktuOptions.contains(satuanWaktu)) satuanWaktu = 'Hari';
    int? productId = p['jp_id'] ?? p['id'];
    return {
      'id': productId,
      'nama': p['jp_nama']?.toString().trim() ?? 'Produk Tanpa Nama',
      'satuan': satuan,
      'harga': p['jp_harga_per'] ?? 0,
      'lama_pengerjaan': p['jp_lama_pengerjaan'] ?? 0,
      'satuan_waktu': satuanWaktu,
      'keterangan': p['jp_keterangan']?.toString().trim() ?? '',
      'icon_path': p['jp_icon_path'],
      'image_url': p['jp_image_url'],
      'image_file_path': null,
      'web_image_bytes': null,
      'should_delete': false,
    };
  }

  Future<void> _submitForm() async {
    if (!_formKey.currentState!.validate()) return;
    if (_produkList.isEmpty) return _showMsg('Minimal satu produk diperlukan', Colors.red);

    setState(() => _isLoading = true);
    try {
      final uri = Uri.parse('$_baseUrl/layanan/with-products/${widget.layananId}');
      final request = http.MultipartRequest('PUT', uri)
        ..headers['Authorization'] = 'Bearer $_accessToken'
        ..fields['ln_layanan'] = _namaLayananCtrl.text.trim()
        ..fields['ln_prioritas'] = _skorPrioritasCtrl.text.trim()
        ..fields['ln_cuci'] = _selectedProses['Cuci']! ? 'Ya' : 'Tidak'
        ..fields['ln_kering'] = _selectedProses['Kering']! ? 'Ya' : 'Tidak'
        ..fields['ln_setrika'] = _selectedProses['Setrika']! ? 'Ya' : 'Tidak';

      List<Map<String, dynamic>> productsToSave = [];
      int fileIdx = 0;

      for (var product in _produkList) {
        if (product['should_delete'] == true) continue;
        final productData = {
          'jp_nama': product['nama'],
          'jp_satuan': product['satuan'],
          'jp_harga_per': product['harga'],
          'jp_lama_pengerjaan': product['lama_pengerjaan'],
          'jp_satuan_waktu': product['satuan_waktu'],
          'jp_keterangan': product['keterangan'],
        };
        if (product['id'] != null) productData['jp_id'] = product['id'];
        if (product['icon_path'] != null) productData['jp_icon_path'] = product['icon_path'];
        if (product['image_url'] != null) productData['jp_image_url'] = product['image_url'];

        productsToSave.add(productData);

        if (product['web_image_bytes'] != null || product['image_file_path'] != null) {
          final key = 'produk_$fileIdx';
          if (kIsWeb) {
            request.files.add(http.MultipartFile.fromBytes(key, product['web_image_bytes'], filename: 'img.jpg', contentType: MediaType('image', 'jpeg')));
          } else {
            request.files.add(await http.MultipartFile.fromPath(key, product['image_file_path'], contentType: MediaType('image', 'jpeg')));
          }
        }
        fileIdx++;
      }

      request.fields['jenis_produk'] = json.encode(productsToSave);
      final response = await http.Response.fromStream(await request.send());

      if (!mounted) return;
      setState(() => _isLoading = false);

      if (response.statusCode == 200) {
        _showMsg('Berhasil diperbarui', Colors.green);
        Navigator.pop(context, true);
      } else {
        _showMsg('Gagal menyimpan data', Colors.red);
      }
    } catch (e) {
      if (mounted) setState(() => _isLoading = false);
      _showMsg('Error: $e', Colors.red);
    }
  }

  void _addEditProduct([int? index]) async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => ModalTambahProduk(
        initialData: index != null ? _produkList[index] : null,
        baseUrl: _baseUrl,
      ),
    );
    if (result != null) {
      setState(() {
        if (index != null) {
          _produkList[index] = result;
        } else {
          result['id'] = null;
          _produkList.add(result);
        }
      });
    }
  }

  void _showMsg(String msg, Color color) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg), backgroundColor: color, behavior: SnackBarBehavior.floating));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Edit Layanan'), backgroundColor: Colors.white, foregroundColor: Colors.black87, elevation: 0),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator(color: Color(0xFFFF6B6B)))
          : Column(
              children: [
                Expanded(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(24),
                    child: Form(
                      key: _formKey,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _buildTextField('Nama Layanan', _namaLayananCtrl),
                          _buildTextField('Skor Prioritas', _skorPrioritasCtrl, isNumber: true),
                          const Text('Proses Termasuk', style: TextStyle(fontWeight: FontWeight.bold)),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceAround,
                            children: _selectedProses.keys.map((k) => Column(children: [Checkbox(value: _selectedProses[k], onChanged: (v) => setState(() => _selectedProses[k] = v!), activeColor: const Color(0xFFFF6B6B)), Text(k)])).toList(),
                          ),
                          const Divider(height: 32),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text('Jenis Produk', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                              IconButton(onPressed: () => _addEditProduct(), icon: const Icon(Icons.add_circle, color: Color(0xFFFF6B6B))),
                            ],
                          ),
                          ListView.builder(
                            shrinkWrap: true,
                            physics: const NeverScrollableScrollPhysics(),
                            itemCount: _produkList.length,
                            itemBuilder: (context, i) {
                              final p = _produkList[i];
                              return Card(
                                margin: const EdgeInsets.only(bottom: 10),
                                child: ListTile(
                                  title: Text(p['nama']),
                                  subtitle: Text('Rp ${currencyFormatter.format(p['harga'])} / ${p['satuan']}'),
                                  trailing: Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(icon: const Icon(Icons.edit, color: Color(0xFFFF6B6B)), onPressed: () => _addEditProduct(i)),
                                      IconButton(icon: const Icon(Icons.delete, color: Colors.red), onPressed: () => setState(() => _produkList.removeAt(i))),
                                    ],
                                  ),
                                ),
                              );
                            },
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.all(16),
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _submitForm,
                    style: ElevatedButton.styleFrom(backgroundColor: const Color(0xFFFF6B6B), padding: const EdgeInsets.symmetric(vertical: 16)),
                    child: const Text('SIMPAN PERUBAHAN', style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold)),
                  ),
                ),
              ],
            ),
    );
  }

  Widget _buildTextField(String label, TextEditingController ctrl, {bool isNumber = false}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: const TextStyle(fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          TextFormField(
            controller: ctrl,
            keyboardType: isNumber ? TextInputType.number : TextInputType.text,
            inputFormatters: isNumber ? [FilteringTextInputFormatter.digitsOnly] : null,
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.grey[50],
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            ),
            validator: (v) => v == null || v.isEmpty ? 'Wajib diisi' : null,
          ),
        ],
      ),
    );
  }
}