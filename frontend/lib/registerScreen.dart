import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'createOutletScreen.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _phoneController = TextEditingController();
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _referralController = TextEditingController();

  bool _isPasswordVisible = false;
  bool _isConfirmPasswordVisible = false;
  bool _agreeToTerms = false;
  bool _isLoading = false;
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
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    _phoneController.dispose();
    _confirmPasswordController.dispose();
    _referralController.dispose();
    super.dispose();
  }

  Future<void> _handleNext() async {
    if (_formKey.currentState!.validate()) {
      if (!_agreeToTerms) {
        _showSnackBar('Harap setujui syarat dan ketentuan', isError: true);
        return;
      }

      setState(() => _isLoading = true);

      try {
        final Map<String, dynamic> requestBody = {
          'username': _nameController.text.trim(),
          'email': _emailController.text.trim().toLowerCase(),
          'nomor_hp': _phoneController.text.trim(),
          'password': _passwordController.text,
          'confirmPassword': _confirmPasswordController.text,
          'group': 'owner',
          'agreeTerms': _agreeToTerms,
          'subscribeNewsletter': false,
        };

        if (_referralController.text.trim().isNotEmpty) {
          requestBody['referralCode'] = _referralController.text
              .trim()
              .toUpperCase();
        }

        if (_selectedSource != null && _selectedSource!.isNotEmpty) {
          requestBody['source'] = _selectedSource;
        }

        print('=== REGISTER REQUEST ===');
        print('URL: $baseUrl/auth/register');
        print('Request Body: ${jsonEncode(requestBody)}');
        print('========================');

        final registerResponse = await http
            .post(
              Uri.parse('$baseUrl/auth/register'),
              headers: {'Content-Type': 'application/json'},
              body: jsonEncode(requestBody),
            )
            .timeout(
              const Duration(seconds: 30),
              onTimeout: () {
                throw Exception('Request timeout');
              },
            );

        print('=== REGISTER RESPONSE ===');
        print('Status Code: ${registerResponse.statusCode}');
        print('Response Headers: ${registerResponse.headers}');
        print('Response Body: ${registerResponse.body}');
        print('=========================');

        if (!mounted) return;

        setState(() => _isLoading = false);

        if (registerResponse.statusCode != 200 &&
            registerResponse.statusCode != 201) {
          try {
            final registerData = jsonDecode(registerResponse.body);
            String errorMessage =
                registerData['error'] ??
                registerData['message'] ??
                'Pendaftaran gagal. Status: ${registerResponse.statusCode}';
            _showSnackBar(errorMessage, isError: true);
          } catch (e) {
            _showSnackBar(
              'Pendaftaran gagal. Status: ${registerResponse.statusCode}',
              isError: true,
            );
          }
          return;
        }
        final registerData = jsonDecode(registerResponse.body);
        print('=== PARSED RESPONSE ===');
        print('Success: ${registerData['success']}');
        print('Message: ${registerData['message']}');
        print('Response Keys: ${registerData.keys.toList()}');
        print('=======================');

        // TAMBAHKAN PARSING VARIABLE INI:
        String? token = registerData['access_token'] as String?;
        String? refreshToken = registerData['refresh_token'] as String?;
        int? userId = registerData['user_id'] as int?; // TAMBAHKAN BARIS INI

        print('=== TOKEN VALIDATION ===');
        print('Access Token found: ${token != null}');
        print('Access Token is empty: ${token?.isEmpty ?? true}');
        print('Access Token length: ${token?.length ?? 0}');
        print('Refresh Token found: ${refreshToken != null}');
        print('User ID: $userId'); 
        if (token != null && token.isNotEmpty) {
          print(
            'Token (first 50 chars): ${token.substring(0, token.length > 50 ? 50 : token.length)}',
          );
        }
        print('========================');

        if (token == null || token.isEmpty) {
          print('ERROR: Access token not found in response');
          print('Available keys: ${registerData.keys.join(", ")}');
          _showSnackBar('Token tidak ditemukan dari server', isError: true);
          return;
        }

      

        if (!mounted) return;

        Navigator.pushReplacement(
          context,
          MaterialPageRoute(
            builder: (context) => CreateOutletScreen(
              token: token!,
              registrationData: {
                'username': _nameController.text.trim(),
                'email': _emailController.text.trim().toLowerCase(),
                'nomor_hp': _phoneController.text.trim(),
                'referralCode': _referralController.text.trim().isNotEmpty
                    ? _referralController.text.trim().toUpperCase()
                    : null,
                'userId': userId?.toString(),
                'access_token': token,
                'refresh_token': refreshToken,
              },
            ),
          ),
        );
      } on SocketException catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        print('ERROR: SocketException - ${e.message}');
        _showSnackBar(
          'Tidak dapat terhubung ke server.\n'
          'Pastikan backend berjalan di: $baseUrl\n'
          'Error: ${e.message}',
          isError: true,
        );
      } on FormatException catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        print('ERROR: FormatException - ${e.message}');
        _showSnackBar(
          'Response dari server tidak valid: ${e.message}',
          isError: true,
        );
      } on Exception catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        print('ERROR: Exception - ${e.toString()}');
        _showSnackBar('Terjadi kesalahan: ${e.toString()}', isError: true);
      } catch (e) {
        if (!mounted) return;
        setState(() => _isLoading = false);
        print('ERROR: Unknown - ${e.toString()}');
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
                    'Selamat datang di ayo cici!',
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
                    'Terima kasih sudah mendownload aplikasi ayo cici.',
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

                        // Email Field
                        _buildTextField(
                          controller: _emailController,
                          hintText: 'Email',
                          icon: Icons.email_outlined,
                          keyboardType: TextInputType.emailAddress,
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'Email tidak boleh kosong';
                            }
                            if (!RegExp(
                              r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$',
                            ).hasMatch(value)) {
                              return 'Email tidak valid';
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

                        // Password Field
                        _buildTextField(
                          controller: _passwordController,
                          hintText: 'Kata sandi',
                          icon: Icons.lock_outline_rounded,
                          isPassword: true,
                          isPasswordVisible: _isPasswordVisible,
                          onTogglePassword: () {
                            setState(() {
                              _isPasswordVisible = !_isPasswordVisible;
                            });
                          },
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'Password tidak boleh kosong';
                            }
                            if (value.length < 8) {
                              return 'Password minimal 8 karakter';
                            }
                            return null;
                          },
                        ),
                        const SizedBox(height: 18),

                        // Confirm Password Field
                        _buildTextField(
                          controller: _confirmPasswordController,
                          hintText: 'Ulangi kata sandi',
                          icon: Icons.lock_outline_rounded,
                          isPassword: true,
                          isPasswordVisible: _isConfirmPasswordVisible,
                          onTogglePassword: () {
                            setState(() {
                              _isConfirmPasswordVisible =
                                  !_isConfirmPasswordVisible;
                            });
                          },
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return 'Konfirmasi password tidak boleh kosong';
                            }
                            if (value != _passwordController.text) {
                              return 'Password tidak cocok';
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
                                onPressed: _isLoading ? null : _handleNext,
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

  Widget _buildTextField({
    required TextEditingController controller,
    required String hintText,
    required IconData icon,
    TextInputType? keyboardType,
    TextCapitalization textCapitalization = TextCapitalization.none,
    bool isPassword = false,
    bool isPasswordVisible = false,
    VoidCallback? onTogglePassword,
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
        obscureText: isPassword && !isPasswordVisible,
        keyboardType: keyboardType,
        textCapitalization: textCapitalization,
        enabled: !_isLoading,
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
          suffixIcon: isPassword
              ? IconButton(
                  icon: Icon(
                    isPasswordVisible
                        ? Icons.visibility_off_rounded
                        : Icons.visibility_rounded,
                    color: Colors.grey[400],
                  ),
                  onPressed: onTogglePassword,
                )
              : null,
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
        validator: validator,
      ),
    );
  }
}
