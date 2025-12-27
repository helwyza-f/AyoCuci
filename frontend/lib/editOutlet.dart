import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'SettingMetodePembayaran.dart';

class EditOutletPage extends StatefulWidget {
  final int outletId;
  final String namaOutlet;
  final String alamat;
  final String nomorHP;
  final String? photo;

  const EditOutletPage({
    super.key,
    required this.outletId,
    required this.namaOutlet,
    required this.alamat,
    required this.nomorHP,
    this.photo,
  });

  @override
  State<EditOutletPage> createState() => _EditOutletPageState();
}

class _EditOutletPageState extends State<EditOutletPage> {
  final _formKey = GlobalKey<FormState>();
  final _namaOutletController = TextEditingController();
  final _alamatController = TextEditingController();
  final _nomorHPController = TextEditingController();

  bool _isLoading = false;
  dynamic _selectedImage;
  final ImagePicker _picker = ImagePicker();

  // Base URL untuk API
  final String _baseUrl = 'http://localhost:8080/v1';
  String? _accessToken;
  String? _initialImageUrl;

  static const Color primaryColor = Color(0xFFEF5350);
  static const Color accentColor = Color(0xFFE91E63);

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _accessToken = prefs.getString('access_token');
      _namaOutletController.text = widget.namaOutlet;
      _alamatController.text = widget.alamat;
      _nomorHPController.text = widget.nomorHP;
      _initialImageUrl = widget.photo;
    });
  }

  Future<void> _pickImage() async {
    try {
      final XFile? image = await _picker.pickImage(
        source: ImageSource.gallery,
        imageQuality: 80,
      );

      if (image != null) {
        setState(() {
          if (kIsWeb) {
            _selectedImage = image;
          } else {
            _selectedImage = File(image.path);
          }
          _initialImageUrl = null;
        });

        _showSuccessSnackBar('Foto berhasil dipilih');
      }
    } catch (e) {
      print('Error picking image: $e');
      _showErrorSnackBar('Gagal memilih foto: ${e.toString()}');
    }
  }

  Future<void> _takePhoto() async {
    try {
      final XFile? photo = await _picker.pickImage(
        source: ImageSource.camera,
        imageQuality: 80,
      );

      if (photo != null) {
        setState(() {
          if (kIsWeb) {
            _selectedImage = photo;
          } else {
            _selectedImage = File(photo.path);
          }
          _initialImageUrl =
              null; // Clear initial image when new photo is taken
        });

        _showSuccessSnackBar('Foto berhasil diambil');
      }
    } catch (e) {
      print('Error taking photo: $e');
      _showErrorSnackBar('Gagal mengambil foto: ${e.toString()}');
    }
  }

  void _showImageSourceDialog() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (BuildContext context) {
        return Container(
          padding: const EdgeInsets.all(20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                  color: Colors.grey[300],
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              const SizedBox(height: 20),
              const Text(
                'Pilih Sumber Foto',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 20),
              ListTile(
                leading: Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: primaryColor.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: const Icon(Icons.photo_library, color: primaryColor),
                ),
                title: const Text('Galeri'),
                onTap: () {
                  Navigator.pop(context);
                  _pickImage();
                },
              ),
              if (!kIsWeb)
                ListTile(
                  leading: Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: primaryColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Icon(Icons.camera_alt, color: primaryColor),
                  ),
                  title: const Text('Kamera'),
                  onTap: () {
                    Navigator.pop(context);
                    _takePhoto();
                  },
                ),
              if (_selectedImage != null || _initialImageUrl != null)
                ListTile(
                  leading: Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: Colors.red.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: const Icon(Icons.delete_outline, color: Colors.red),
                  ),
                  title: const Text('Hapus Gambar'),
                  onTap: () {
                    setState(() {
                      _selectedImage = null;
                      _initialImageUrl = null;
                    });
                    Navigator.pop(context);
                  },
                ),
              const SizedBox(height: 10),
            ],
          ),
        );
      },
    );
  }

  Future<void> _saveOutlet() async {
    if (_formKey.currentState!.validate()) {
      setState(() => _isLoading = true);

      try {
        var uri = Uri.parse('$_baseUrl/outlet/${widget.outletId}');
        var request = http.MultipartRequest('PUT', uri);

        request.headers['Authorization'] = 'Bearer $_accessToken';
        request.fields['nama_outlet'] = _namaOutletController.text.trim();
        request.fields['alamat'] = _alamatController.text.trim();
        request.fields['nomor_hp'] = _nomorHPController.text.trim();

        if (_selectedImage != null) {
          if (kIsWeb) {
            final bytes = await (_selectedImage as XFile).readAsBytes();
            request.files.add(
              http.MultipartFile.fromBytes(
                'photo',
                bytes,
                filename: (_selectedImage as XFile).name,
              ),
            );
          } else {
            request.files.add(
              await http.MultipartFile.fromPath(
                'photo',
                (_selectedImage as File).path,
              ),
            );
          }
        }

        var response = await request.send();
        var responseBody = await response.stream.bytesToString();
        print("Response: $responseBody");

        if (response.statusCode == 200) {
          _showSuccessSnackBar('Outlet berhasil diperbarui');
          Navigator.pop(context, true);
        } else {
          _showErrorSnackBar('Error: $responseBody');
        }
      } catch (e) {
        _showErrorSnackBar("Error: $e");
      } finally {
        setState(() => _isLoading = false);
      }
    }
  }

  void _showSuccessSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.check_circle, color: Colors.white),
            const SizedBox(width: 12),
            Text(message),
          ],
        ),
        backgroundColor: Colors.green,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 2),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  void _showErrorSnackBar(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.error_outline, color: Colors.white),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: Colors.red,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 3),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        margin: const EdgeInsets.all(16),
      ),
    );
  }

  @override
  void dispose() {
    _namaOutletController.dispose();
    _alamatController.dispose();
    _nomorHPController.dispose();
    super.dispose();
  }

  Widget _buildMenuCard({
    required String title,
    required String subtitle,
    required IconData icon,
    required VoidCallback onTap,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFFFE4E4), width: 2),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(16),
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        title,
                        style: const TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: Colors.black87,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        subtitle,
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey[600],
                          fontWeight: FontWeight.w400,
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: Colors.grey[100],
                    shape: BoxShape.circle,
                  ),
                  child: Icon(icon, size: 18, color: Colors.black87),
                ),
                const SizedBox(width: 8),
                Container(
                  width: 36,
                  height: 36,
                  decoration: const BoxDecoration(
                    color: Colors.black87,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.arrow_forward_ios,
                    size: 14,
                    color: Colors.white,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildImagePreview() {
    // Sama seperti EditLayananPage, remove /v1 dari base URL
    final baseServer = _baseUrl.replaceAll('/v1', '');
    Widget imageWidget;

    if (_selectedImage != null) {
      // New image selected
      if (kIsWeb) {
        imageWidget = FutureBuilder<Uint8List>(
          future: (_selectedImage as XFile).readAsBytes(),
          builder: (context, snapshot) {
            if (snapshot.hasData) {
              return Image.memory(
                snapshot.data!,
                fit: BoxFit.cover,
                width: 100,
                height: 100,
              );
            }
            return const CircularProgressIndicator();
          },
        );
      } else {
        imageWidget = Image.file(
          _selectedImage as File,
          fit: BoxFit.cover,
          width: 100,
          height: 100,
        );
      }
    } else if (_initialImageUrl != null && _initialImageUrl!.isNotEmpty) {
      // Display initial image from server
      final fullImageUrl = _initialImageUrl!.startsWith('http')
          ? _initialImageUrl!
          : '$baseServer/$_initialImageUrl';

      imageWidget = Image.network(
        fullImageUrl,
        fit: BoxFit.cover,
        errorBuilder: (context, error, stackTrace) {
          print('Error loading image: $error');
          print('Image URL: $fullImageUrl');
          return const Icon(Icons.broken_image, size: 50, color: Colors.grey);
        },
      );
    } else {
      // No image
      imageWidget = const Icon(Icons.store, size: 50, color: Colors.grey);
    }

    return Container(
      width: 100,
      height: 100,
      decoration: BoxDecoration(
        color: Colors.grey[200],
        shape: BoxShape.circle,
      ),
      child: ClipOval(child: imageWidget),
    );
  }

  void _showEditProfileDialog() {
    showDialog(
      context: context,
      builder: (context) => Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        backgroundColor: Colors.white,
        child: Container(
          constraints: const BoxConstraints(maxWidth: 500),
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
                      children: [
                        const Expanded(
                          child: Text(
                            'Edit Profile Outlet',
                            style: TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.bold,
                              color: Colors.black87,
                            ),
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
                    const SizedBox(height: 24),

                    // Photo Section
                    Center(
                      child: Stack(
                        children: [
                          _buildImagePreview(),
                          Positioned(
                            bottom: 0,
                            right: 0,
                            child: GestureDetector(
                              onTap: _showImageSourceDialog,
                              child: Container(
                                padding: const EdgeInsets.all(8),
                                decoration: BoxDecoration(
                                  color: primaryColor,
                                  shape: BoxShape.circle,
                                  border: Border.all(
                                    color: Colors.white,
                                    width: 2,
                                  ),
                                ),
                                child: const Icon(
                                  Icons.camera_alt,
                                  size: 16,
                                  color: Colors.white,
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 24),

                    // Nama Outlet
                    const Text(
                      'Nama Outlet',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Colors.black87,
                      ),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _namaOutletController,
                      decoration: InputDecoration(
                        hintText: 'Masukkan nama outlet',
                        hintStyle: TextStyle(
                          color: Colors.grey[400],
                          fontSize: 14,
                        ),
                        filled: true,
                        fillColor: const Color(0xFFF5F5F5),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide.none,
                        ),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 14,
                        ),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Nama outlet tidak boleh kosong';
                        }
                        if (value.trim().length < 3) {
                          return 'Nama outlet minimal 3 karakter';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),

                    // Alamat
                    const Text(
                      'Alamat',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Colors.black87,
                      ),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _alamatController,
                      maxLines: 3,
                      decoration: InputDecoration(
                        hintText: 'Masukkan alamat lengkap',
                        hintStyle: TextStyle(
                          color: Colors.grey[400],
                          fontSize: 14,
                        ),
                        filled: true,
                        fillColor: const Color(0xFFF5F5F5),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide.none,
                        ),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 14,
                        ),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'Alamat tidak boleh kosong';
                        }
                        if (value.trim().length < 10) {
                          return 'Alamat minimal 10 karakter';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'No. Handphone',
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: Colors.black87,
                      ),
                    ),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _nomorHPController,
                      keyboardType: TextInputType.phone,
                      decoration: InputDecoration(
                        hintText: '081234567890',
                        hintStyle: TextStyle(
                          color: Colors.grey[400],
                          fontSize: 14,
                        ),
                        filled: true,
                        fillColor: const Color(0xFFF5F5F5),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide.none,
                        ),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 14,
                        ),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return 'No Handphone tidak boleh kosong';
                        }
                        if (value.trim().length < 10) {
                          return 'No Handphone minimal 10 digit';
                        }
                        if (value.trim().length > 15) {
                          return 'No Handphone maksimal 15 digit';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 24),

                    // Save Button
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: ElevatedButton(
                        onPressed: _isLoading
                            ? null
                            : () async {
                                await _saveOutlet();
                                if (mounted) Navigator.pop(context);
                              },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: primaryColor,
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
                                'Simpan Perubahan',
                                style: TextStyle(
                                  fontSize: 15,
                                  fontWeight: FontWeight.w600,
                                  color: Colors.white,
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
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8F8F8),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black87),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          'Konfigurasi Profil Outlet',
          style: TextStyle(
            color: Colors.black87,
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      body: Stack(
        children: [
          ListView(
            padding: const EdgeInsets.all(16),
            children: [
              _buildMenuCard(
                title: 'Edit Profile Outlet',
                subtitle: 'Edit Profile Outlet anda disini',
                icon: Icons.settings,
                onTap: _showEditProfileDialog,
              ),
              _buildMenuCard(
                title: 'Pengaturan Outlet',
                subtitle: 'Atur Pengaturan Outlet anda disini',
                icon: Icons.settings,
                onTap: () {
                  // Handle pengaturan outlet
                },
              ),
              _buildMenuCard(
                title: 'Pengaturan Metode Pembayaran',
                subtitle: 'Atur Metode Pembayaran disini',
                icon: Icons.settings,
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => PaymentMethodPage(
                        outletId: widget.outletId,
                      ),
                    ),
                  );
                },
              ),
            ],
          ),
          if (_isLoading)
            Container(
              color: Colors.black.withOpacity(0.3),
              child: const Center(
                child: CircularProgressIndicator(color: primaryColor),
              ),
            ),
        ],
      ),
    );
  }
}
