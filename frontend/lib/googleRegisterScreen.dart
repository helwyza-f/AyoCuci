import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'createOutletScreen.dart';

class GoogleRegisterScreen extends StatefulWidget {
  final String googleId;
  final String email;
  final String? displayName;
  final String? photoUrl;
  final String? idToken;
  final String? accessToken;

  const GoogleRegisterScreen({
    super.key,
    required this.googleId,
    required this.email,
    this.displayName,
    this.photoUrl,
    this.idToken,
    this.accessToken,
  });

  @override
  State<GoogleRegisterScreen> createState() => _GoogleRegisterScreenState();
}

class _GoogleRegisterScreenState extends State<GoogleRegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  final _referralController = TextEditingController();

  bool _isLoading = false;
  bool _agreeToTerms = false;
  String? _selectedSource;

  final List<Map<String, dynamic>> _sourceOptions = [
    {'icon': '📱', 'label': 'Instagram', 'value': 'instagram'},
    {'icon': '🎵', 'label': 'Tiktok', 'value': 'tiktok'},
    {'icon': '👥', 'label': 'Facebook', 'value': 'facebook'},
    {
      'icon': '💬',
      'label': 'Direkomendasikan oleh teman/saudara',
      'value': 'referral',
    },
    {'icon': '🔍', 'label': 'Google Search', 'value': 'google'},
    {'icon': '📺', 'label': 'Youtube', 'value': 'youtube'},
    {'icon': '🌐', 'label': 'Lainnya', 'value': 'others'},
  ];

  @override
  void initState() {
    super.initState();

    if (widget.displayName != null && widget.displayName!.isNotEmpty) {
      _nameController.text = widget.displayName!;
    }

    if (kDebugMode) {
      print('=== GoogleRegisterScreen Initialized ===');
      print('Google ID: ${widget.googleId}');
      print('Email: ${widget.email}');
      print('Display Name: ${widget.displayName}');
      print('Has ID Token: ${widget.idToken != null}');
      print('Has Access Token: ${widget.accessToken != null}');
      print('=======================================');
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    _referralController.dispose();
    super.dispose();
  }

  Future<void> _handleRegister() async {
    if (!_formKey.currentState!.validate()) return;

    if (!_agreeToTerms) {
      _showSnackBar('Harap setujui syarat dan ketentuan', isError: true);
      return;
    }

    setState(() => _isLoading = true);

    try {
      if (widget.idToken == null && widget.accessToken == null) {
        throw Exception('Token Google tidak tersedia. Silakan login ulang.');
      }

      final requestBody = {
        'google_id': widget.googleId,
        'email': widget.email,
        'username': _nameController.text.trim(),
        'nomor_hp': _phoneController.text.trim(),
        'group': 'owner',
        'agreeTerms': _agreeToTerms,
        'subscribeNewsletter': false,
      };

      if (_referralController.text.trim().isNotEmpty) {
        requestBody['referralCode'] =
            _referralController.text.trim().toUpperCase();
      }

      if (_selectedSource != null && _selectedSource!.isNotEmpty) {
        requestBody['source'] = _selectedSource!;
      }

      if (kDebugMode) {
        print('=== Google Register Request ===');
        print('Request Body: ${jsonEncode(requestBody)}');
        print('==============================');
      }

      final response = await http.post(
        Uri.parse('http://localhost:8080/v1/auth/google/register'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode(requestBody),
      );

      if (kDebugMode) {
        print('=== Register Response ===');
        print('Status Code: ${response.statusCode}');
        print('Response Body: ${response.body}');
        print('Content-Type: ${response.headers['content-type']}');
        print('========================');
      }

      if (!mounted) return;

      if (response.body.isEmpty) {
        throw Exception('Server mengembalikan response kosong');
      }

      final contentType = response.headers['content-type'];
      if (contentType != null && !contentType.contains('application/json')) {
        throw Exception('Response bukan JSON: $contentType');
      }

      final result = jsonDecode(response.body);

      if (response.statusCode == 200 || response.statusCode == 201) {
        final accessToken = result['data']?['access_token'];
        final refreshToken = result['data']?['refresh_token'];
        final userData = result['data']?['user'];

        if (kDebugMode) {
          print('=== Token Validation ===');
          print('Access Token found: ${accessToken != null}');
          print('Refresh Token found: ${refreshToken != null}');
          print('User Data: $userData');
          print('=======================');
        }

        if (accessToken == null || accessToken.isEmpty) {
          throw Exception('Access token tidak ditemukan dalam response');
        }

        _showSnackBar('Registrasi berhasil! Selamat datang!', isError: false);

        if (!mounted) return;

        Navigator.pushReplacement(
          context,
          MaterialPageRoute(
            builder: (context) => CreateOutletScreen(
              token: accessToken,
              registrationData: {
                'username': _nameController.text.trim(),
                'email': widget.email,
                'nomor_hp': _phoneController.text.trim(),
                'nama_lengkap':
                    userData?['username'] ?? _nameController.text.trim(),
                'user_id': userData?['id']?.toString() ?? '',
                'referralCode': _referralController.text.trim().isNotEmpty
                    ? _referralController.text.trim().toUpperCase()
                    : null,
                'source': _selectedSource,
              },
            ),
          ),
        );
      } else if (response.statusCode == 409) {
        _showSnackBar(
          result['message'] ?? 'Email atau nomor HP sudah terdaftar',
          isError: true,
        );
      } else if (response.statusCode == 401) {
        _showSnackBar('Token Google tidak valid. Silakan login ulang.', isError: true);
      } else {
        _showSnackBar(
          result['error'] ?? result['message'] ?? 'Registrasi gagal',
          isError: true,
        );
      }
    } on FormatException catch (e) {
      if (kDebugMode) print('FormatException: $e');
      if (!mounted) return;
      _showSnackBar('Format response tidak valid dari server', isError: true);
    } on http.ClientException catch (e) {
      if (kDebugMode) print('ClientException: $e');
      if (!mounted) return;
      _showSnackBar('Tidak dapat terhubung ke server', isError: true);
    } catch (e) {
      if (kDebugMode) print('Error: $e');
      if (!mounted) return;
      _showSnackBar('Error: ${e.toString()}', isError: true);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _showReferralInfo() {
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return Dialog(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          child: Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(20),
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFF5F4E).withOpacity(0.1),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.help_outline_rounded,
                    size: 40,
                    color: Color(0xFFFF5F4E),
                  ),
                ),
                const SizedBox(height: 20),
                const Text(
                  'Apa itu Kode Referral?',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF2C3E50),
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  'Punya kode referral? Masukkan kode tersebut ketika anda mendaftar untuk mendapatkan bonus dan keuntungan tambahan!\n\nFormat kode: REFxxxxxx (8-20 karakter)',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 14,
                    color: Colors.grey[600],
                    height: 1.5,
                  ),
                ),
                const SizedBox(height: 24),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: () => Navigator.pop(context),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: const Color(0xFFFF5F4E),
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      elevation: 0,
                    ),
                    child: const Text(
                      'Mengerti',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
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

  Widget _buildTextField({
    required TextEditingController controller,
    required String hintText,
    required IconData icon,
    TextInputType? keyboardType,
    TextCapitalization textCapitalization = TextCapitalization.none,
    String? Function(String?)? validator,
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
        inputFormatters: keyboardType == TextInputType.phone
            ? [FilteringTextInputFormatter.digitsOnly]
            : null,
        decoration: InputDecoration(
          hintText: hintText,
          hintStyle: TextStyle(
            color: Colors.grey[400],
            fontSize: 14,
            fontWeight: FontWeight.w400,
          ),
          prefixIcon: Container(
            margin: const EdgeInsets.only(left: 12, right: 8),
            child: Icon(
              icon,
              color: const Color(0xFFFF5F4E),
              size: 22,
            ),
          ),
          prefixIconConstraints: const BoxConstraints(
            minWidth: 50,
          ),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: BorderSide(
              color: Colors.grey[200]!,
            ),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: BorderSide(
              color: Colors.grey[200]!,
              width: 1.5,
            ),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(
              color: Color(0xFFFF5F4E),
              width: 2,
            ),
          ),
          errorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(
              color: Colors.red,
              width: 1.5,
            ),
          ),
          focusedErrorBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(16),
            borderSide: const BorderSide(
              color: Colors.red,
              width: 2,
            ),
          ),
          filled: true,
          fillColor: Colors.white,
          contentPadding:
              const EdgeInsets.symmetric(horizontal: 20, vertical: 18),
        ),
        validator: validator,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFFF5F4E),
      body: SafeArea(
        child: Column(
          children: [
            // Stepper Progress
            Container(
              width: double.infinity,
              margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 20),
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
                  _buildStepPill(1, 'Data Akun', true),
                  const SizedBox(width: 8),
                  _buildStepPill(2, 'Data Outlet', false),
                ],
              ),
            ),

            // Welcome Message
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
                    'Lengkapi profil Anda',
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
                    'Masukkan informasi tambahan untuk melengkapi pendaftaran',
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

            // Form Container
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
                        const SizedBox(height: 8),

                        // Email Field (Read-only with verified badge)
                        Container(
                          padding: const EdgeInsets.all(18),
                          decoration: BoxDecoration(
                            color: Colors.grey[50],
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: Colors.green.withOpacity(0.3),
                              width: 1.5,
                            ),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.grey.withOpacity(0.08),
                                blurRadius: 12,
                                offset: const Offset(0, 4),
                              ),
                            ],
                          ),
                          child: Row(
                            children: [
                              Container(
                                margin: const EdgeInsets.only(right: 12),
                                child: const Icon(
                                  Icons.email_outlined,
                                  color: Color(0xFFFF5F4E),
                                  size: 22,
                                ),
                              ),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Email (Terverifikasi)',
                                      style: TextStyle(
                                        fontSize: 11,
                                        color: Colors.grey[600],
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                    const SizedBox(height: 4),
                                    Text(
                                      widget.email,
                                      style: const TextStyle(
                                        fontSize: 14,
                                        color: Color(0xFF2C3E50),
                                        fontWeight: FontWeight.w600,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Icon(
                                Icons.verified,
                                color: Colors.green[600],
                                size: 24,
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 18),

                        // Name Field
                        _buildTextField(
                          controller: _nameController,
                          hintText: 'Nama lengkap anda',
                          icon: Icons.person_outline_rounded,
                          keyboardType: TextInputType.name,
                          textCapitalization: TextCapitalization.words,
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'Nama tidak boleh kosong';
                            }
                            if (value.length < 3) {
                              return 'Nama minimal 3 karakter';
                            }
                            if (value.length > 100) {
                              return 'Nama maksimal 100 karakter';
                            }
                            return null;
                          },
                        ),
                        const SizedBox(height: 18),

                        // Phone Field
                        _buildTextField(
                          controller: _phoneController,
                          hintText: 'No handphone owner (Whatsapp aktif)',
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

                        // Kode Referal (Optional) with Info Button
                        Row(
                          children: [
                            Expanded(
                              child: _buildTextField(
                                controller: _referralController,
                                hintText: 'Kode referal (opsional)',
                                icon: Icons.card_giftcard_rounded,
                                keyboardType: TextInputType.text,
                                textCapitalization:
                                    TextCapitalization.characters,
                                validator: (value) {
                                  if (value != null && value.isNotEmpty) {
                                    if (value.length < 8 || value.length > 20) {
                                      return 'Kode referral 8-20 karakter';
                                    }
                                    if (!value.toUpperCase().startsWith(
                                      'REF',
                                    )) {
                                      return 'Kode harus dimulai dengan REF';
                                    }
                                  }
                                  return null;
                                },
                              ),
                            ),
                            const SizedBox(width: 8),
                            GestureDetector(
                              onTap: () => _showReferralInfo(),
                              child: Container(
                                width: 50,
                                height: 56,
                                decoration: BoxDecoration(
                                  color: Colors.grey[100],
                                  borderRadius: BorderRadius.circular(16),
                                  border: Border.all(
                                    color: Colors.grey[300]!,
                                    width: 1.5,
                                  ),
                                ),
                                child: Icon(
                                  Icons.help_outline_rounded,
                                  color: Colors.grey[600],
                                  size: 24,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 18),

                        // Darimana tau aplikasi - Dropdown
                        Container(
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
                          child: DropdownButtonFormField<String>(
                            value: _selectedSource,
                            decoration: InputDecoration(
                              hintText:
                                  'Darimana anda tau aplikasi ini? (opsional)',
                              hintStyle: TextStyle(
                                color: Colors.grey[400],
                                fontSize: 14,
                                fontWeight: FontWeight.w400,
                              ),
                              prefixIcon: Container(
                                margin: const EdgeInsets.only(
                                  left: 12,
                                  right: 8,
                                ),
                                child: const Icon(
                                  Icons.public_rounded,
                                  color: Color(0xFFFF5F4E),
                                  size: 22,
                                ),
                              ),
                              prefixIconConstraints: const BoxConstraints(
                                minWidth: 50,
                              ),
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(16),
                                borderSide: BorderSide(
                                  color: Colors.grey[200]!,
                                ),
                              ),
                              enabledBorder: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(16),
                                borderSide: BorderSide(
                                  color: Colors.grey[200]!,
                                  width: 1.5,
                                ),
                              ),
                              focusedBorder: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(16),
                                borderSide: const BorderSide(
                                  color: Color(0xFFFF5F4E),
                                  width: 2,
                                ),
                              ),
                              filled: true,
                              fillColor: Colors.white,
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 20,
                                vertical: 18,
                              ),
                            ),
                            items: _sourceOptions.map((option) {
                              return DropdownMenuItem<String>(
                                value: option['value'],
                                child: Row(
                                  children: [
                                    Text(
                                      option['icon'],
                                      style: const TextStyle(fontSize: 18),
                                    ),
                                    const SizedBox(width: 12),
                                    Flexible(
                                      child: Text(
                                        option['label'],
                                        style: const TextStyle(
                                          fontSize: 14,
                                          fontWeight: FontWeight.w500,
                                          color: Color(0xFF2C3E50),
                                        ),
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            }).toList(),
                            onChanged: _isLoading
                                ? null
                                : (value) {
                                    setState(() {
                                      _selectedSource = value;
                                    });
                                  },
                            icon: Icon(
                              Icons.keyboard_arrow_down_rounded,
                              color: Colors.grey[400],
                            ),
                            dropdownColor: Colors.white,
                            isExpanded: true,
                          ),
                        ),
                        const SizedBox(height: 24),

                        // Terms Checkbox
                        Container(
                          padding: const EdgeInsets.all(14),
                          decoration: BoxDecoration(
                            color: Colors.grey[50],
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: _agreeToTerms
                                  ? const Color(0xFFFF5F4E).withOpacity(0.3)
                                  : Colors.grey[200]!,
                              width: 1.5,
                            ),
                          ),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.center,
                            children: [
                              SizedBox(
                                height: 24,
                                width: 24,
                                child: Checkbox(
                                  value: _agreeToTerms,
                                  onChanged: _isLoading
                                      ? null
                                      : (value) {
                                          setState(() {
                                            _agreeToTerms = value ?? false;
                                          });
                                        },
                                  activeColor: const Color(0xFFFF5F4E),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                ),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: RichText(
                                  text: TextSpan(
                                    style: TextStyle(
                                      color: Colors.grey[700],
                                      fontSize: 13,
                                      height: 1.4,
                                    ),
                                    children: const [
                                      TextSpan(
                                        text: 'Dengan mendaftar saya setuju ',
                                      ),
                                      TextSpan(
                                        text: 'Syarat dan Ketentuan',
                                        style: TextStyle(
                                          color: Color(0xFFFF5F4E),
                                          fontWeight: FontWeight.w600,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 32),

                        // Action Buttons
                        Row(
                          children: [
                            Expanded(
                              child: OutlinedButton(
                                onPressed: _isLoading
                                    ? null
                                    : () {
                                        Navigator.pop(context);
                                      },
                                style: OutlinedButton.styleFrom(
                                  side: BorderSide(
                                    color: Colors.grey[300]!,
                                    width: 1.5,
                                  ),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(30),
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 16,
                                  ),
                                  elevation: 0,
                                ),
                                child: Text(
                                  'KEMBALI',
                                  style: TextStyle(
                                    color: Colors.grey[600],
                                    fontSize: 14,
                                    fontWeight: FontWeight.w700,
                                    letterSpacing: 0.8,
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(width: 14),
                            Expanded(
                              child: ElevatedButton(
                                onPressed: _isLoading ? null : _handleRegister,
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: const Color(0xFFFF5F4E),
                                  foregroundColor: Colors.white,
                                  disabledBackgroundColor: Colors.grey[300],
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(30),
                                  ),
                                  padding: const EdgeInsets.symmetric(
                                    vertical: 16,
                                  ),
                                  elevation: 4,
                                  shadowColor: const Color(
                                    0xFFFF5F4E,
                                  ).withOpacity(0.4),
                                ),
                                child: _isLoading
                                    ? const SizedBox(
                                        height: 20,
                                        width: 20,
                                        child: CircularProgressIndicator(
                                          color: Colors.white,
                                          strokeWidth: 2.5,
                                        ),
                                      )
                                    : const Text(
                                        'SELANJUTNYA',
                                        style: TextStyle(
                                          fontSize: 14,
                                          fontWeight: FontWeight.w700,
                                          letterSpacing: 0.8,
                                        ),
                                      ),
                              ),
                            ),
                          ],
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
    );
  }
}