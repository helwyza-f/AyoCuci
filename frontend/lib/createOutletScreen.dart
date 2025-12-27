import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';
import 'loginScreen.dart';

class Region {
  final String id;
  final String name;

  Region({required this.id, required this.name});

  factory Region.fromJson(Map<String, dynamic> json) {
    return Region(
      id: json['id']?.toString() ?? json['province_id']?.toString() ?? '',
      name: json['name'] as String,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Region && runtimeType == other.runtimeType && id == other.id;

  @override
  int get hashCode => id.hashCode;

  @override
  String toString() => name;
}

class LocationService {
  static const baseUrl = 'https://www.emsifa.com/api-wilayah-indonesia/api';

  static Future<List<Region>> getProvinces() async {
    final res = await http.get(Uri.parse('$baseUrl/provinces.json'));
    if (res.statusCode != 200) return [];

    final data = jsonDecode(res.body) as List;
    return data
        .map((e) => Region(id: e['id'].toString(), name: e['name']))
        .toList();
  }

  static Future<List<Region>> getCities(String provId) async {
    final res = await http.get(Uri.parse('$baseUrl/regencies/$provId.json'));
    if (res.statusCode != 200) return [];

    final data = jsonDecode(res.body) as List;
    return data
        .map((e) => Region(id: e['id'].toString(), name: e['name']))
        .toList();
  }

  static Future<List<Region>> getDistricts(String cityId) async {
    final res = await http.get(Uri.parse('$baseUrl/districts/$cityId.json'));
    if (res.statusCode != 200) return [];

    final data = jsonDecode(res.body) as List;
    return data
        .map((e) => Region(id: e['id'].toString(), name: e['name']))
        .toList();
  }
}

class CreateOutletScreen extends StatefulWidget {
  final String token;
  final Map<String, dynamic> registrationData;

  const CreateOutletScreen({
    super.key,
    required this.token,
    required this.registrationData,
  });

  @override
  State<CreateOutletScreen> createState() => _CreateOutletScreenState();
}

class _CreateOutletScreenState extends State<CreateOutletScreen> {
  final _formKey = GlobalKey<FormState>();
  final _outletNameController = TextEditingController();
  final _outletAddressController = TextEditingController();
  final _outletPhoneController = TextEditingController();

  bool _isLoading = false;
  String? _cachedToken;

  final ImagePicker _picker = ImagePicker();
  XFile? _selectedImage;

  List<Region> _provinces = [];
  List<Region> _cities = [];
  List<Region> _districts = [];

  Region? _selectedProvince;
  Region? _selectedCity;
  Region? _selectedDistrict;

  bool _isLoadingProvinces = false;
  bool _isLoadingCities = false;
  bool _isLoadingDistricts = false;

  static String get baseUrl {
    if (kIsWeb) {
      return 'http://localhost:8080/v1';
    } else if (Platform.isAndroid) {
      return 'http://10.0.2.2:8080/v1';
    } else if (Platform.isIOS) {
      return 'http://localhost:8080/v1';
    } else {
      return 'http://localhost:8080/v1';
    }
  }

  @override
  void initState() {
    super.initState();
    _cachedToken = widget.token;
    _loadProvinces();
  }

  @override
  void dispose() {
    _outletNameController.dispose();
    _outletAddressController.dispose();
    _outletPhoneController.dispose();
    super.dispose();
  }

  Future<void> _loadProvinces() async {
    if (_isLoadingProvinces) return;
    setState(() => _isLoadingProvinces = true);
    try {
      final provinces = await LocationService.getProvinces();
      setState(() {
        _provinces = provinces;
        _isLoadingProvinces = false;
      });
    } catch (e) {
      setState(() => _isLoadingProvinces = false);
      _showSnackBar(
        'Gagal memuat data provinsi: ${e.toString()}',
        isError: true,
      );
    }
  }

  Future<void> _loadCities(String provinceId) async {
    if (_isLoadingCities) return;
    setState(() {
      _isLoadingCities = true;
      _cities = [];
      _districts = [];
      _selectedCity = null;
      _selectedDistrict = null;
    });

    try {
      final cities = await LocationService.getCities(provinceId);
      setState(() {
        _cities = cities;
        _isLoadingCities = false;
      });
    } catch (e) {
      setState(() => _isLoadingCities = false);
      _showSnackBar('Gagal memuat data kota: ${e.toString()}', isError: true);
    }
  }

  Future<void> _loadDistricts(String cityId) async {
    if (_isLoadingDistricts) return;
    setState(() {
      _isLoadingDistricts = true;
      _districts = [];
      _selectedDistrict = null;
    });

    try {
      final districts = await LocationService.getDistricts(cityId);
      setState(() {
        _districts = districts;
        _isLoadingDistricts = false;
      });
    } catch (e) {
      setState(() => _isLoadingDistricts = false);
      _showSnackBar(
        'Gagal memuat data kecamatan: ${e.toString()}',
        isError: true,
      );
    }
  }

  Future<void> _pickImage() async {
    try {
      final XFile? image = await _picker.pickImage(
        source: ImageSource.gallery,
        maxWidth: 1920,
        maxHeight: 1080,
        imageQuality: 85,
      );

      if (image != null) {
        setState(() {
          _selectedImage = image;
        });
      }
    } catch (e) {
      _showSnackBar('Gagal memilih gambar: ${e.toString()}', isError: true);
    }
  }

  Future<void> _removeImage() async {
    setState(() {
      _selectedImage = null;
    });
  }

  Future<void> _handleCompleteRegistration() async {
    if (_formKey.currentState!.validate()) {
      if (_selectedProvince == null) {
        _showSnackBar('Silakan pilih provinsi', isError: true);
        return;
      }
      if (_selectedCity == null) {
        _showSnackBar('Silakan pilih kota', isError: true);
        return;
      }
      if (_selectedDistrict == null) {
        _showSnackBar('Silakan pilih kecamatan', isError: true);
        return;
      }

      if (_cachedToken == null || _cachedToken!.isEmpty) {
        _showSnackBar(
          'Token tidak valid. Silakan ulangi pendaftaran.',
          isError: true,
        );
        return;
      }

      setState(() => _isLoading = true);

      try {
        var request = http.MultipartRequest(
          'POST',
          Uri.parse('$baseUrl/outlet'),
        );

        final cleanToken = _cachedToken!.trim();
        request.headers['Authorization'] = 'Bearer $cleanToken';
        request.headers['Content-Type'] = 'multipart/form-data';
        request.fields['nama_outlet'] = _outletNameController.text.trim();
        request.fields['alamat'] = _outletAddressController.text.trim();
        request.fields['nomor_hp'] = _outletPhoneController.text.trim();
        request.fields['provinsi'] = _selectedProvince!.name;
        request.fields['kota'] = _selectedCity!.name;
        request.fields['kecamatan'] = _selectedDistrict!.name;
        if (_selectedImage != null) {
          if (kIsWeb) {
            var bytes = await _selectedImage!.readAsBytes();
            request.files.add(
              http.MultipartFile.fromBytes(
                'photo',
                bytes,
                filename: _selectedImage!.name,
              ),
            );
          } else {
            request.files.add(
              await http.MultipartFile.fromPath(
                'photo',
                _selectedImage!.path,
                filename: _selectedImage!.name,
              ),
            );
          }
        }

        var streamedResponse = await request.send().timeout(
          const Duration(seconds: 30),
          onTimeout: () {
            throw Exception('Request timeout. Pastikan backend berjalan.');
          },
        );

        var outletResponse = await http.Response.fromStream(streamedResponse);

        if (!mounted) return;

        final outletData = jsonDecode(outletResponse.body);

        if (outletResponse.statusCode == 200 ||
            outletResponse.statusCode == 201) {
          setState(() => _isLoading = false);

          _showSnackBar(
            'Pendaftaran dan outlet berhasil dibuat!',
            isError: false,
          );

          await Future.delayed(const Duration(seconds: 1));

          if (!mounted) return;

          Navigator.pushAndRemoveUntil(
            context,
            MaterialPageRoute(builder: (context) => const LoginScreen()),
            (route) => false,
          );
        } else if (outletResponse.statusCode == 401) {
          setState(() => _isLoading = false);
          _showSnackBar(
            'Token tidak valid atau kadaluarsa. Silakan login kembali.',
            isError: true,
          );

          await Future.delayed(const Duration(seconds: 2));

          if (!mounted) return;

          Navigator.pushAndRemoveUntil(
            context,
            MaterialPageRoute(builder: (context) => const LoginScreen()),
            (route) => false,
          );
        } else {
          setState(() => _isLoading = false);
          String errorMessage =
              outletData['message'] ??
              outletData['error'] ??
              'Gagal membuat outlet. Status: ${outletResponse.statusCode}';
          _showSnackBar(errorMessage, isError: true);
        }
      } on TimeoutException {
        if (!mounted) return;
        setState(() => _isLoading = false);
        _showSnackBar(
          'Request timeout. Pastikan backend berjalan di $baseUrl',
          isError: true,
        );
      } on SocketException catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        _showSnackBar(
          'Tidak dapat terhubung ke server.\n'
          'Pastikan backend berjalan di: $baseUrl\n'
          'Error: ${e.message}',
          isError: true,
        );
      } on FormatException catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        _showSnackBar(
          'Response dari server tidak valid: ${e.message}',
          isError: true,
        );
      } catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        _showSnackBar('Terjadi kesalahan: ${e.toString()}', isError: true);
      }
    }
  }

  void _showSnackBar(String message, {required bool isError}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: isError ? const Color(0xFFFF5F4E) : Colors.green,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 4),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String hintText,
    required IconData icon,
    required TextInputType keyboardType,
    TextCapitalization textCapitalization = TextCapitalization.none,
    String? Function(String?)? validator,
    int maxLines = 1,
  }) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.08),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: TextFormField(
        controller: controller,
        keyboardType: keyboardType,
        textCapitalization: textCapitalization,
        enabled: !_isLoading,
        maxLines: maxLines,
        style: const TextStyle(
          fontSize: 14,
          fontWeight: FontWeight.w500,
          color: Color(0xFF2C3E50),
        ),
        decoration: InputDecoration(
          hintText: hintText,
          hintStyle: TextStyle(
            color: Colors.grey[400],
            fontSize: 14,
            fontWeight: FontWeight.w400,
          ),
          prefixIcon: Container(
            margin: const EdgeInsets.only(left: 12, right: 8),
            child: Icon(icon, color: const Color(0xFFFF5F4E), size: 22),
          ),
          prefixIconConstraints: const BoxConstraints(minWidth: 50),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: BorderSide(color: Colors.grey[200]!),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: BorderSide(color: Colors.grey[200]!, width: 1.5),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(color: Color(0xFFFF5F4E), width: 2),
          ),
          errorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(color: Color(0xFFFF5F4E), width: 1.5),
          ),
          focusedErrorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(color: Color(0xFFFF5F4E), width: 2),
          ),
          filled: true,
          fillColor: Colors.white,
          contentPadding: EdgeInsets.symmetric(
            horizontal: 20,
            vertical: maxLines > 1 ? 16 : 18,
          ),
        ),
        validator: validator,
      ),
    );
  }

  Widget _buildStepPill(int step, String label, bool isActive) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        decoration: BoxDecoration(
          color: isActive ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(30),
          border: isActive
              ? null
              : Border.all(color: Colors.white.withOpacity(0.5), width: 1),
          boxShadow: isActive
              ? [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 5,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircleAvatar(
              radius: 10,
              backgroundColor: isActive
                  ? const Color(0xFFFF5F4E)
                  : Colors.white,
              child: Text(
                step.toString(),
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: isActive ? Colors.white : const Color(0xFFFF5F4E),
                ),
              ),
            ),
            const SizedBox(width: 8),
            Flexible(
              child: Text(
                label,
                style: TextStyle(
                  color: isActive ? const Color(0xFF2C3E50) : Colors.white,
                  fontSize: 12,
                  fontWeight: isActive ? FontWeight.w700 : FontWeight.w500,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildImageUploadSection() {
    return Column(
      children: [
        GestureDetector(
          onTap: _isLoading ? null : _pickImage,
          child: _selectedImage == null
              ? Container(
                  height: 180,
                  decoration: BoxDecoration(
                    color: Colors.grey[50],
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: Colors.grey[300]!,
                      width: 1.5,
                      style: BorderStyle.solid,
                    ),
                  ),
                  child: Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: const Color(0xFFFF5F4E).withOpacity(0.1),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(
                            Icons.add_photo_alternate_outlined,
                            size: 40,
                            color: Color(0xFFFF5F4E),
                          ),
                        ),
                        const SizedBox(height: 12),
                        Text(
                          'Upload Foto Outlet',
                          style: TextStyle(
                            color: Colors.grey[700],
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '(Opsional)',
                          style: TextStyle(
                            color: Colors.grey[500],
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ),
                )
              : Stack(
                  children: [
                    Container(
                      height: 180,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                          color: const Color(0xFFFF5F4E),
                          width: 2,
                        ),
                      ),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(14),
                        child: kIsWeb
                            ? Image.network(
                                _selectedImage!.path,
                                width: double.infinity,
                                height: 180,
                                fit: BoxFit.cover,
                              )
                            : Image.file(
                                File(_selectedImage!.path),
                                width: double.infinity,
                                height: 180,
                                fit: BoxFit.cover,
                              ),
                      ),
                    ),
                    Positioned(
                      top: 8,
                      right: 8,
                      child: GestureDetector(
                        onTap: _removeImage,
                        child: Container(
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: Colors.black.withOpacity(0.6),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(
                            Icons.close,
                            color: Colors.white,
                            size: 18,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
        ),
        if (_selectedImage == null)
          const SizedBox(height: 8)
        else
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: TextButton.icon(
              onPressed: _isLoading ? null : _pickImage,
              icon: const Icon(Icons.edit, color: Color(0xFFFF5F4E), size: 18),
              label: const Text(
                'Ganti Foto',
                style: TextStyle(
                  color: Color(0xFFFF5F4E),
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                ),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildLocationDropdown({
    required String label,
    required IconData icon,
    required List<Region> items,
    required Region? value,
    required Function(Region?) onChanged,
    required bool isLoading,
    String? Function(Region?)? validator,
  }) {
    return Container(
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.08),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: DropdownButtonFormField<Region>(
        value: value,
        decoration: InputDecoration(
          hintText: label,
          hintStyle: TextStyle(
            color: Colors.grey[400],
            fontSize: 14,
            fontWeight: FontWeight.w400,
          ),
          prefixIcon: Container(
            margin: const EdgeInsets.only(left: 12, right: 8),
            child: Icon(icon, color: const Color(0xFFFF5F4E), size: 22),
          ),
          prefixIconConstraints: const BoxConstraints(minWidth: 50),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: BorderSide(color: Colors.grey[200]!),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: BorderSide(color: Colors.grey[200]!, width: 1.5),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(color: Color(0xFFFF5F4E), width: 2),
          ),
          errorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(color: Color(0xFFFF5F4E), width: 1.5),
          ),
          focusedErrorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(color: Color(0xFFFF5F4E), width: 2),
          ),
          filled: true,
          fillColor: Colors.white,
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 20,
            vertical: 18,
          ),
        ),
        items: isLoading
            ? []
            : items.map((region) {
                return DropdownMenuItem<Region>(
                  value: region,
                  child: Text(
                    region.name,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                      color: Color(0xFF2C3E50),
                    ),
                  ),
                );
              }).toList(),
        onChanged: _isLoading ? null : onChanged,
        validator: validator,
        isExpanded: true,
        icon: isLoading
            ? const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  valueColor: AlwaysStoppedAnimation<Color>(Color(0xFFFF5F4E)),
                ),
              )
            : const Icon(Icons.arrow_drop_down, color: Color(0xFFFF5F4E)),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return WillPopScope(
      onWillPop: () async => false,
      child: Scaffold(
        backgroundColor: const Color(0xFFFF5F4E),
        body: SafeArea(
          child: Column(
            children: [
              Container(
                width: double.infinity,
                margin: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 20,
                ),
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.25),
                  borderRadius: BorderRadius.circular(35),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.1),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    _buildStepPill(1, 'Data Akun', false),
                    const SizedBox(width: 8),
                    _buildStepPill(2, 'Data Outlet', true),
                  ],
                ),
              ),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.only(
                  top: 4,
                  bottom: 24,
                  left: 28,
                  right: 28,
                ),
                child: const Column(
                  children: [
                    Text(
                      'Lengkapi Data Outlet',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.w700,
                        letterSpacing: 0.3,
                      ),
                    ),
                    SizedBox(height: 8),
                    Text(
                      'Masukkan informasi outlet Anda untuk melanjutkan',
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 13,
                        fontWeight: FontWeight.w400,
                        height: 1.4,
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: Container(
                  width: double.infinity,
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: const BorderRadius.only(
                      topLeft: Radius.circular(35),
                      topRight: Radius.circular(35),
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.1),
                        blurRadius: 20,
                        offset: const Offset(0, -5),
                      ),
                    ],
                  ),
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.fromLTRB(28, 50, 28, 28),
                    child: Form(
                      key: _formKey,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _buildImageUploadSection(),
                          const SizedBox(height: 24),

                          _buildTextField(
                            controller: _outletNameController,
                            hintText: 'Nama outlet',
                            icon: Icons.store_outlined,
                            keyboardType: TextInputType.text,
                            textCapitalization: TextCapitalization.words,
                            validator: (value) {
                              if (value == null || value.isEmpty) {
                                return 'Nama outlet tidak boleh kosong';
                              }
                              if (value.length < 3) {
                                return 'Nama outlet minimal 3 karakter';
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: 18),
                          _buildTextField(
                            controller: _outletPhoneController,
                            hintText: 'No handphone outlet (Whatsapp aktif)',
                            icon: Icons.phone_android_rounded,
                            keyboardType: TextInputType.phone,
                            validator: (value) {
                              if (value == null || value.isEmpty) {
                                return 'Nomor HP tidak boleh kosong';
                              }
                              if (!RegExp(r'^[0-9]+$').hasMatch(value)) {
                                return 'Nomor HP harus berupa angka';
                              }
                              if (value.length < 10) {
                                return 'Nomor HP minimal 10 digit';
                              }
                              if (value.length > 15) {
                                return 'Nomor HP maksimal 15 digit';
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: 18),
                          _buildLocationDropdown(
                            label: 'Pilih Provinsi',
                            icon: Icons.map_outlined,
                            items: _provinces,
                            value: _selectedProvince,
                            isLoading: _isLoadingProvinces,
                            onChanged: (Region? newValue) {
                              setState(() {
                                _selectedProvince = newValue;
                                _selectedCity = null;
                                _selectedDistrict = null;
                                _cities = [];
                                _districts = [];
                              });
                              if (newValue != null) {
                                _loadCities(newValue.id);
                              }
                            },
                            validator: (value) {
                              if (value == null) {
                                return 'Provinsi wajib dipilih';
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: 18),
                          Row(
                            children: [
                              Expanded(
                                child: _buildLocationDropdown(
                                  label: 'Pilih Kota',
                                  icon: Icons.location_city_outlined,
                                  items: _cities,
                                  value: _selectedCity,
                                  isLoading: _isLoadingCities,
                                  onChanged: (Region? newValue) {
                                    setState(() {
                                      _selectedCity = newValue;
                                      _selectedDistrict = null;
                                      _districts = [];
                                    });
                                    if (newValue != null) {
                                      _loadDistricts(newValue.id);
                                    }
                                  },
                                  validator: (value) {
                                    if (value == null) {
                                      return 'Kota wajib dipilih';
                                    }
                                    return null;
                                  },
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: _buildLocationDropdown(
                                  label: 'Pilih Kecamatan',
                                  icon: Icons.location_on_outlined,
                                  items: _districts,
                                  value: _selectedDistrict,
                                  isLoading: _isLoadingDistricts,
                                  onChanged: (Region? newValue) {
                                    setState(() {
                                      _selectedDistrict = newValue;
                                    });
                                  },
                                  validator: (value) {
                                    if (value == null) {
                                      return 'Kecamatan wajib dipilih';
                                    }
                                    return null;
                                  },
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 18),
                          _buildTextField(
                            controller: _outletAddressController,
                            hintText: 'Alamat lengkap (maks. 255 karakter)',
                            icon: Icons.home_outlined,
                            keyboardType: TextInputType.streetAddress,
                            textCapitalization: TextCapitalization.sentences,
                            maxLines: 3,
                            validator: (value) {
                              if (value == null || value.isEmpty) {
                                return 'Alamat tidak boleh kosong';
                              }
                              if (value.length < 10) {
                                return 'Alamat minimal 10 karakter';
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: 30),

                          // Action Button
                          SizedBox(
                            width: double.infinity,
                            height: 55,
                            child: ElevatedButton(
                              onPressed: _isLoading
                                  ? null
                                  : _handleCompleteRegistration,
                              style: ElevatedButton.styleFrom(
                                backgroundColor: const Color(0xFFFF5F4E),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(16),
                                ),
                                elevation: 0,
                              ),
                              child: _isLoading
                                  ? const CircularProgressIndicator(
                                      color: Colors.white,
                                      strokeWidth: 3,
                                    )
                                  : const Text(
                                      'Selesaikan Pendaftaran',
                                      style: TextStyle(
                                        color: Colors.white,
                                        fontSize: 16,
                                        fontWeight: FontWeight.w700,
                                      ),
                                    ),
                            ),
                          ),
                          const SizedBox(height: 20),
                        ],
                      ),
                    ),
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
