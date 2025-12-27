import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'registerScreen.dart';
import 'homeScreen.dart';
import 'forgopasswordScreen.dart';
import 'googleRegisterScreen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen>
    with SingleTickerProviderStateMixin {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _isPasswordVisible = false;
  bool _isLoading = false;
  bool _rememberMe = false;

  late AnimationController _animationController;
  late Animation<double> _fadeAnimation;
  late Animation<Offset> _slideAnimation;
  late GoogleSignIn _googleSignIn;

  final String _baseUrl = 'http://localhost:8080/v1';

  @override
  void initState() {
    super.initState();
    _initializeGoogleSignIn();
    _initializeAnimations();
    _loadRememberMe();
  }

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _animationController.dispose();
    super.dispose();
  }

  void _initializeGoogleSignIn() {
    final String? googleClientId = dotenv.env['GOOGLE_WEB_CLIENT_ID'];

    if (googleClientId == null || googleClientId.isEmpty) {
      debugPrint("ERROR: GOOGLE_WEB_CLIENT_ID not found in .env");
    }

    _googleSignIn = GoogleSignIn(
      clientId: kIsWeb ? googleClientId : null,
      scopes: ['email', 'profile', 'openid'],
      serverClientId: kIsWeb ? null : googleClientId,
    );

    if (kDebugMode) {
      print('Platform: ${kIsWeb ? "WEB" : "MOBILE"}');
      print('Client ID configured: ${googleClientId != null}');
    }
  }

  void _initializeAnimations() {
    _animationController = AnimationController(
      duration: const Duration(milliseconds: 1000),
      vsync: this,
    );

    _fadeAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeInOut),
    );

    _slideAnimation =
        Tween<Offset>(begin: const Offset(0, 0.3), end: Offset.zero).animate(
          CurvedAnimation(
            parent: _animationController,
            curve: Curves.easeOutCubic,
          ),
        );

    _animationController.forward();
  }

  Future<void> _saveUserData(Map<String, dynamic> userData) async {
    final prefs = await SharedPreferences.getInstance();

    final userId =
        userData['user_id']?.toString() ?? userData['id']?.toString() ?? '';

    await prefs.setString('user_id', userId);
    await prefs.setString('username', userData['username']?.toString() ?? '');
    await prefs.setString('email', userData['email']?.toString() ?? '');
    await prefs.setString('nomor_hp', userData['nomor_hp']?.toString() ?? '');
    await prefs.setString('group', userData['group']?.toString() ?? '');
    await prefs.setBool('has_outlet', userData['has_outlet'] ?? false);
    await prefs.setBool('isLoggedIn', true);

    if (userData['outlet_id'] != null) {
      final outletId = int.tryParse(userData['outlet_id'].toString());
      if (outletId != null) {
        await prefs.setInt('outletId', outletId);
        await prefs.setInt('outlet_id', outletId);

        if (kDebugMode) {
          print('Outlet ID saved with both keys: $outletId');
        }
      }
    } else {
      await prefs.remove('outletId');
      await prefs.remove('outlet_id');
      if (kDebugMode) {
        print('No outlet_id in response');
      }
    }

    if (kDebugMode) {
      print('=== User Data Saved ===');
      print('User ID: $userId');
      print('Username: ${userData['username']}');
      print('Email: ${userData['email']}');
      print('Nomor HP: ${userData['nomor_hp']}');
      print('Group: ${userData['group']}');
      print('Has Outlet: ${userData['has_outlet']}');
      print('Outlet ID: ${userData['outlet_id']}');
      print('======================');
    }
  }

  Future<void> _saveTokens(String accessToken, String refreshToken) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('access_token', accessToken);
    await prefs.setString('refresh_token', refreshToken);

    if (kDebugMode) {
      print('Tokens saved successfully');
      print('Access Token: ${accessToken.substring(0, 20)}...');
      print('Refresh Token: ${refreshToken.substring(0, 20)}...');
    }
  }

  Future<void> _saveCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    if (_rememberMe) {
      await prefs.setString('saved_email', _emailController.text.trim());
      await prefs.setBool('remember_me', true);
    } else {
      await prefs.remove('saved_email');
      await prefs.setBool('remember_me', false);
    }
  }

  Future<void> _loadRememberMe() async {
    final prefs = await SharedPreferences.getInstance();
    final savedEmail = prefs.getString('saved_email');
    final rememberMe = prefs.getBool('remember_me') ?? false;

    if (mounted && rememberMe && savedEmail != null) {
      setState(() {
        _emailController.text = savedEmail;
        _rememberMe = rememberMe;
      });
    }
  }

  Future<void> _handleLogin() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    try {
      final response = await http.post(
        Uri.parse('$_baseUrl/auth/login'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'email': _emailController.text.trim(),
          'password': _passwordController.text,
          'remember_me': _rememberMe ? 'y' : 'n',
        }),
      );

      if (kDebugMode) {
        print('=== Login Response ===');
        print('Status: ${response.statusCode}');
        print('Body: ${response.body}');
        print('=====================');
      }

      if (!mounted) return;

      if (response.statusCode == 200) {
        await _handleLoginSuccess(response);
      } else {
        _handleLoginError(response);
      }
    } on http.ClientException catch (e) {
      if (kDebugMode) print('Network error: $e');
      _showErrorSnackBar(
        'Tidak dapat terhubung ke server. Pastikan backend sedang berjalan.',
      );
    } catch (e) {
      if (kDebugMode) print('Login error: $e');
      if (mounted) {
        _showErrorSnackBar(_getErrorMessage(e));
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _handleLoginSuccess(http.Response response) async {
    try {
      final result = jsonDecode(response.body);

      if (kDebugMode) {
        print('=== Parsed Login Response ===');
        print('Full Response: $result');
        print('============================');
      }

      final data = result['data'];

      if (data == null) {
        _showErrorSnackBar('Format response tidak valid');
        return;
      }

      final accessToken = data['access_token']?.toString();
      final refreshToken = data['refresh_token']?.toString();

      if (accessToken == null || accessToken.isEmpty) {
        _showErrorSnackBar('Token akses tidak ditemukan');
        return;
      }

      if (refreshToken == null || refreshToken.isEmpty) {
        _showErrorSnackBar('Refresh token tidak ditemukan');
        return;
      }

      final userId = data['user_id']?.toString();
      if (userId == null || userId.isEmpty) {
        _showErrorSnackBar('User ID tidak ditemukan');
        return;
      }

      await _saveTokens(accessToken, refreshToken);
      await _saveUserData(data);
      await _saveCredentials();

      if (kDebugMode) {
        print('✓ Login successful!');
      }

      _showSuccessSnackBar('Login berhasil!');
      await Future.delayed(const Duration(milliseconds: 500));

      if (mounted) {
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => const LaundryHomePage()),
        );
      }
    } catch (e) {
      if (kDebugMode) {
        print('Error parsing login response: $e');
      }
      _showErrorSnackBar('Gagal memproses response dari server');
    }
  }

  void _handleLoginError(http.Response response) {
    String errorMessage = 'Gagal login';
    int? remainingAttempts;
    int? remainingSeconds;

    try {
      final result = jsonDecode(response.body);
      errorMessage = result['message'] ?? result['error'] ?? errorMessage;
      remainingAttempts = result['remaining_attempts'];
      remainingSeconds = result['remaining_seconds'];

      if (kDebugMode) {
        print('Error Response: $result');
      }
    } catch (e) {
      if (kDebugMode) print('Error parsing error response: $e');
    }
    switch (response.statusCode) {
      case 401:
        if (remainingAttempts != null) {
          errorMessage =
              'Email atau password salah. Sisa percobaan: $remainingAttempts';
        } else {
          errorMessage = 'Email atau password salah';
        }
        break;
      case 403:
        errorMessage = 'Akun tidak aktif';
        break;
      case 404:
        errorMessage = 'Akun tidak ditemukan';
        break;
      case 429:
        if (remainingSeconds != null) {
          final minutes = (remainingSeconds / 60).ceil();
          errorMessage =
              'Terlalu banyak percobaan gagal. Coba lagi dalam $minutes menit';
        }
        break;
    }

    _showErrorSnackBar(errorMessage);
  }

  Future<void> _handleGoogleSignIn() async {
    if (_isLoading) return;

    setState(() => _isLoading = true);

    try {
      await _googleSignIn.signOut();

      final GoogleSignInAccount? googleUser = await _googleSignIn.signIn();

      if (googleUser == null) {
        if (kDebugMode) print('User cancelled sign-in');
        return;
      }

      final GoogleSignInAuthentication googleAuth =
          await googleUser.authentication;
      final String? accessToken = googleAuth.accessToken;

      if (kDebugMode) {
        print('=== Google Sign-In ===');
        print('Email: ${googleUser.email}');
        print('Google ID: ${googleUser.id}');
        print('Token: ${accessToken != null ? "✓" : "✗"}');
        print('=====================');
      }

      if (accessToken == null) {
        throw Exception('No access token from Google');
      }

      final response = await http.post(
        Uri.parse('$_baseUrl/auth/google/signin'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({
          'google_id': googleUser.id,
          'email': googleUser.email,
          'name': googleUser.displayName,
          'photoUrl': googleUser.photoUrl,
          'accessToken': accessToken,
        }),
      );

      if (!mounted) return;

      await _handleGoogleSignInResponse(response, googleUser, accessToken);
    } catch (e) {
      if (kDebugMode) print('Google Sign-In error: $e');
      if (!mounted) return;
      _showErrorSnackBar(_getErrorMessage(e));
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _handleGoogleSignInResponse(
    http.Response response,
    GoogleSignInAccount googleUser,
    String accessToken,
  ) async {
    if (kDebugMode) {
      print('=== Backend Response ===');
      print('Status: ${response.statusCode}');
      print('Body: ${response.body}');
      print('=======================');
    }

    if (response.body.isEmpty) {
      _showErrorSnackBar('Response kosong dari server');
      return;
    }

    dynamic result;
    try {
      result = jsonDecode(response.body);
    } catch (e) {
      if (kDebugMode) print('JSON Parse Error: $e');
      _showErrorSnackBar('Format response tidak valid');
      return;
    }

    switch (response.statusCode) {
      case 200:
        await _handleGoogleLoginSuccess(result);
        break;

      case 404:
        if (kDebugMode) print('User not registered, navigating to register');
        if (mounted) {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => GoogleRegisterScreen(
                googleId: googleUser.id,
                email: googleUser.email,
                photoUrl: googleUser.photoUrl,
                accessToken: accessToken,
              ),
            ),
          );
        }
        break;

      case 409:
        _showWarningSnackBar(
          result?['message'] ??
              'Email sudah terdaftar dengan metode login lain',
        );
        break;

      case 401:
        _showErrorSnackBar(result?['message'] ?? 'Token Google tidak valid');
        break;

      default:
        _showErrorSnackBar(
          result?['message'] ??
              'Gagal login dengan Google (${response.statusCode})',
        );
    }
  }

  Future<void> _handleGoogleLoginSuccess(dynamic result) async {
    try {
      final data = result['data'];

      if (data == null) {
        _showErrorSnackBar('Format response tidak valid');
        return;
      }

      final backendAccessToken = data['access_token']?.toString();
      final refreshToken = data['refresh_token']?.toString();

      if (kDebugMode) {
        print('=== Google Login Data ===');
        print('Access Token: ${backendAccessToken != null ? "✓" : "✗"}');
        print('Refresh Token: ${refreshToken != null ? "✓" : "✗"}');
        print('User Data: $data');
        print('========================');
      }

      if (backendAccessToken == null || backendAccessToken.isEmpty) {
        _showErrorSnackBar('Token akses tidak ditemukan');
        return;
      }

      if (refreshToken == null || refreshToken.isEmpty) {
        _showErrorSnackBar('Refresh token tidak ditemukan');
        return;
      }

      final userId = data['user_id']?.toString();
      if (userId == null || userId.isEmpty) {
        _showErrorSnackBar('User ID tidak ditemukan');
        return;
      }
      await _saveTokens(backendAccessToken, refreshToken);
      await _saveUserData(data);
      if (kDebugMode) {
        print('✓ Google login successful!');
      }
      _showSuccessSnackBar('Login berhasil!');
      await Future.delayed(const Duration(milliseconds: 500));

      if (mounted) {
        Navigator.pushReplacement(
          context,
          MaterialPageRoute(builder: (context) => const LaundryHomePage()),
        );
      }
    } catch (e) {
      if (kDebugMode) {
        print('Error parsing Google login response: $e');
      }
      _showErrorSnackBar('Gagal memproses response dari server');
    }
  }

  String _getErrorMessage(dynamic error) {
    final errorStr = error.toString();

    if (errorStr.contains('Failed host lookup') ||
        errorStr.contains('Connection refused')) {
      return 'Server tidak dapat dijangkau. Pastikan backend sedang berjalan.';
    } else if (errorStr.contains('SocketException')) {
      return 'Tidak dapat terhubung ke server. Cek koneksi internet Anda.';
    } else if (errorStr.contains('token')) {
      return 'Gagal mendapatkan token dari Google. Coba lagi.';
    } else if (errorStr.contains('TimeoutException')) {
      return 'Koneksi timeout. Coba lagi.';
    }
    return 'Terjadi kesalahan saat login';
  }

  void _showSuccessSnackBar(String message) {
    _showSnackBar(message, Colors.green, Icons.check_circle);
  }

  void _showWarningSnackBar(String message) {
    _showSnackBar(message, Colors.orange, Icons.warning);
  }

  void _showErrorSnackBar(String message) {
    _showSnackBar(message, Colors.red, Icons.error_outline);
  }

  void _showSnackBar(String message, Color color, IconData icon) {
    if (!mounted) return;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            Icon(icon, color: Colors.white),
            const SizedBox(width: 12),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(10)),
        duration: const Duration(seconds: 3),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    final isSmallScreen = size.height < 700;

    return Scaffold(
      backgroundColor: Colors.white,
      body: SafeArea(
        child: SingleChildScrollView(
          child: ConstrainedBox(
            constraints: BoxConstraints(
              minHeight: size.height - MediaQuery.of(context).padding.vertical,
            ),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 24.0),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  SizedBox(height: isSmallScreen ? 20 : 40),
                  _buildHeader(isSmallScreen, size),
                  SizedBox(height: isSmallScreen ? 30 : 40),
                  _buildLoginForm(),
                  SizedBox(height: isSmallScreen ? 20 : 30),
                  _buildSignUpLink(),
                  SizedBox(height: isSmallScreen ? 20 : 30),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(bool isSmallScreen, Size size) {
    return SlideTransition(
      position: _slideAnimation,
      child: FadeTransition(
        opacity: _fadeAnimation,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            SizedBox(
              height: isSmallScreen ? size.height * 0.10 : size.height * 0.13,
              child: Image.asset(
                'assets/images/logo.png',
                fit: BoxFit.contain,
                errorBuilder: (context, error, stackTrace) => Icon(
                  Icons.local_laundry_service_rounded,
                  size: isSmallScreen ? 60 : 80,
                  color: const Color(0xFFF4593B),
                ),
              ),
            ),
            const SizedBox(height: 24),

            Text(
              'Selamat Datang',
              style: TextStyle(
                fontSize: isSmallScreen ? 26 : 32,
                fontWeight: FontWeight.bold,
                color: const Color(0xFFF4593B),
                letterSpacing: 0.5,
              ),
            ),
            const SizedBox(height: 12),

            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: Text(
                'Senang menyambut Anda kembali. Mari lihat apa yang Anda kerjakan selanjutnya!',
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: isSmallScreen ? 14 : 16,
                  color: Colors.grey[600],
                  height: 1.5,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLoginForm() {
    return SlideTransition(
      position: _slideAnimation,
      child: FadeTransition(
        opacity: _fadeAnimation,
        child: Form(
          key: _formKey,
          child: Column(
            children: [
              _buildEmailField(),
              const SizedBox(height: 16),
              _buildPasswordField(),
              const SizedBox(height: 12),
              _buildRememberMeRow(),
              const SizedBox(height: 24),
              _buildLoginButton(),
              const SizedBox(height: 20),
              _buildDivider(),
              const SizedBox(height: 20),
              _buildGoogleSignInButton(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmailField() {
    return TextFormField(
      controller: _emailController,
      keyboardType: TextInputType.emailAddress,
      enabled: !_isLoading,
      decoration: InputDecoration(
        labelText: 'Email',
        hintText: 'Masukkan email Anda',
        prefixIcon: const Icon(Icons.email_outlined, color: Color(0xFFF4593B)),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey[300]!),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFFF4593B), width: 2),
        ),
        filled: true,
        fillColor: Colors.grey[50],
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
      ),
      validator: (value) {
        if (value == null || value.isEmpty) {
          return 'Email tidak boleh kosong';
        }
        if (!RegExp(r'^[^@]+@[^@]+\.[^@]+').hasMatch(value.trim())) {
          return 'Email tidak valid';
        }
        return null;
      },
    );
  }

  Widget _buildPasswordField() {
    return TextFormField(
      controller: _passwordController,
      obscureText: !_isPasswordVisible,
      enabled: !_isLoading,
      decoration: InputDecoration(
        labelText: 'Password',
        hintText: 'Masukkan password Anda',
        prefixIcon: const Icon(Icons.lock_outline, color: Color(0xFFF4593B)),
        suffixIcon: IconButton(
          icon: Icon(
            _isPasswordVisible
                ? Icons.visibility_outlined
                : Icons.visibility_off_outlined,
            color: Colors.grey[600],
          ),
          onPressed: _isLoading
              ? null
              : () => setState(() => _isPasswordVisible = !_isPasswordVisible),
        ),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: BorderSide(color: Colors.grey[300]!),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFFF4593B), width: 2),
        ),
        filled: true,
        fillColor: Colors.grey[50],
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
      ),
      validator: (value) {
        if (value == null || value.isEmpty) {
          return 'Password tidak boleh kosong';
        }
        if (value.length < 6) {
          return 'Password minimal 6 karakter';
        }
        return null;
      },
    );
  }

  Widget _buildRememberMeRow() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Row(
          children: [
            SizedBox(
              height: 20,
              width: 20,
              child: Checkbox(
                value: _rememberMe,
                onChanged: _isLoading
                    ? null
                    : (value) => setState(() => _rememberMe = value ?? false),
                activeColor: const Color(0xFFF4593B),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(4),
                ),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              'Ingat Saya',
              style: TextStyle(
                color: Colors.grey[700],
                fontSize: 14,
              ),
            ),
          ],
        ),
        TextButton(
          onPressed: _isLoading
              ? null
              : () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ForgotPasswordScreen(),
                    ),
                  );
                },
          style: TextButton.styleFrom(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            minimumSize: const Size(0, 36),
          ),
          child: const Text(
            'Lupa Password?',
            style: TextStyle(
              color: Color(0xFFF4593B),
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildLoginButton() {
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: ElevatedButton(
        onPressed: _isLoading ? null : _handleLogin,
        style: ElevatedButton.styleFrom(
          backgroundColor: const Color(0xFFF4593B),
          foregroundColor: Colors.white,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          elevation: 2,
          shadowColor: const Color(0xFFF4593B).withOpacity(0.3),
          disabledBackgroundColor: Colors.grey[300],
        ),
        child: _isLoading
            ? const SizedBox(
                height: 22,
                width: 22,
                child: CircularProgressIndicator(
                  color: Colors.white,
                  strokeWidth: 2.5,
                ),
              )
            : const Text(
                'Masuk',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 0.5,
                ),
              ),
      ),
    );
  }

  Widget _buildDivider() {
    return Row(
      children: [
        Expanded(child: Divider(color: Colors.grey[300], thickness: 1)),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16),
          child: Text(
            'atau',
            style: TextStyle(
              color: Colors.grey[600],
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
        Expanded(child: Divider(color: Colors.grey[300], thickness: 1)),
      ],
    );
  }

  Widget _buildGoogleSignInButton() {
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: OutlinedButton.icon(
        onPressed: _isLoading ? null : _handleGoogleSignIn,
        style: OutlinedButton.styleFrom(
          side: BorderSide(color: Colors.grey[300]!, width: 1.5),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          backgroundColor: Colors.white,
          disabledBackgroundColor: Colors.grey[50],
        ),
        icon: _isLoading
            ? SizedBox(
                height: 22,
                width: 22,
                child: CircularProgressIndicator(
                  strokeWidth: 2.5,
                  valueColor: AlwaysStoppedAnimation(Colors.grey[400]!),
                ),
              )
            : Image.asset(
                'assets/images/google.png',
                errorBuilder: (context, error, stackTrace) =>
                    const Icon(Icons.g_mobiledata, color: Color(0xFF4285F4), size: 28),
                height: 24,
                width: 24,
              ),
        label: Text(
          _isLoading ? 'Menghubungkan...' : 'Masuk dengan Google',
          style: TextStyle(
            fontSize: 15,
            fontWeight: FontWeight.w600,
            color: _isLoading ? Colors.grey[400] : Colors.black87,
          ),
        ),
      ),
    );
  }

  Widget _buildSignUpLink() {
    return FadeTransition(
      opacity: _fadeAnimation,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            'Belum punya akun? ',
            style: TextStyle(
              color: Colors.grey[700],
              fontSize: 14,
            ),
          ),
          TextButton(
            onPressed: _isLoading
                ? null
                : () => Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const RegisterScreen(),
                      ),
                    ),
            style: TextButton.styleFrom(
              padding: EdgeInsets.zero,
              minimumSize: const Size(0, 0),
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
            child: const Text(
              'Daftar Sekarang',
              style: TextStyle(
                color: Color(0xFFF4593B),
                fontSize: 14,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ],
      ),
    );
  }
}