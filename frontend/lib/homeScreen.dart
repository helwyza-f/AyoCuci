import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'dart:async';
import 'settingScreen.dart';
import 'tambahKaryawan.dart';
import 'bottomNav.dart';
import 'tambahPelanggan.dart';
import 'pengeluaran.dart';
import 'package:intl/intl.dart';

class LaundryHomePage extends StatefulWidget {
  const LaundryHomePage({super.key});

  @override
  State<LaundryHomePage> createState() => _LaundryHomePageState();
}

class _LaundryHomePageState extends State<LaundryHomePage> {
  String _username = '';
  String _email = '';
  bool _isLoading = true;
  int _selectedIndex = 0;
  int _currentInfoPage = 0;
  int _currentPromoPage = 0;

  String _currentTime = '';
  String _currentDate = '';
  Timer? _timeTimer;

  PageController? _infoPageController;
  PageController? _promoPageController;
  Timer? _infoTimer;
  Timer? _promoTimer;

  @override
  void initState() {
    super.initState();
    _infoPageController = PageController();
    _promoPageController = PageController();
    _loadUserData();
    _startAutoSlide();
    _updateTime();
    _startTimeUpdate();
  }

  void _updateTime() {
    setState(() {
      DateTime now = DateTime.now();
      _currentTime = DateFormat('HH:mm:ss').format(now);
      List<String> hari = [
        'Minggu',
        'Senin',
        'Selasa',
        'Rabu',
        'Kamis',
        'Jumat',
        'Sabtu',
      ];
      List<String> bulan = [
        '',
        'Januari',
        'Februari',
        'Maret',
        'April',
        'Mei',
        'Juni',
        'Juli',
        'Agustus',
        'September',
        'Oktober',
        'November',
        'Desember',
      ];

      _currentDate =
          '${hari[now.weekday % 7]}, ${now.day} ${bulan[now.month]} ${now.year}';
    });
  }

  void _startTimeUpdate() {
    _timeTimer = Timer.periodic(Duration(seconds: 1), (timer) {
      _updateTime();
    });
  }

  void _startAutoSlide() {
    _infoTimer = Timer.periodic(Duration(seconds: 3), (timer) {
      if (_infoPageController != null && _infoPageController!.hasClients) {
        int nextPage = (_currentInfoPage + 1) % 3;
        _infoPageController!.animateToPage(
          nextPage,
          duration: Duration(milliseconds: 600),
          curve: Curves.easeInOut,
        );
      }
    });

    _promoTimer = Timer.periodic(Duration(seconds: 8), (timer) {
      if (_promoPageController != null && _promoPageController!.hasClients) {
        int nextPage = (_currentPromoPage + 1) % 3;
        _promoPageController!.animateToPage(
          nextPage,
          duration: Duration(milliseconds: 600),
          curve: Curves.easeInOut,
        );
      }
    });
  }

  @override
  void dispose() {
    _timeTimer?.cancel();
    _infoTimer?.cancel();
    _promoTimer?.cancel();
    _infoPageController?.dispose();
    _promoPageController?.dispose();
    super.dispose();
  }

  Future<void> _loadUserData() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _username = prefs.getString('username') ?? 'Putriyana';
      _email = prefs.getString('email') ?? '';
      _isLoading = false;
    });
  }

  void _onNavItemTapped(int index) {
    if (index == _selectedIndex) return;

    if (index == 1) {
      // Navigator.push(context, MaterialPageRoute(builder: (_) => OrderPage()));
    } else if (index == 2) {
      // Navigator.push(context, MaterialPageRoute(builder: (_) => StatistikPage()));
    } else if (index == 3) {
      Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => SettingPage()),
      ).then((_) {
        setState(() => _selectedIndex = 0);
      });
    } else {
      setState(() => _selectedIndex = index);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Stack(
              children: [
                Container(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        Color(0xFFEF5350),
                        Color(0xFFFF7043),
                        Colors.grey[100]!,
                      ],
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                      stops: [0.0, 0.3, 0.5],
                    ),
                  ),
                ),
                SafeArea(
                  child: Column(
                    children: [
                      _buildHeader(),
                      Expanded(
                        child: SingleChildScrollView(
                          padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const SizedBox(height: 16),
                              _buildMenuGrid(),
                              const SizedBox(height: 24),
                              _buildInformasiSection(),
                              const SizedBox(height: 24),
                              _buildPromoSection(),
                              const SizedBox(height: 80),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
      floatingActionButton: CustomCenterFab(onPressed: () {}),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
      bottomNavigationBar: CustomBottomNav(
        selectedIndex: _selectedIndex,
        onItemTapped: _onNavItemTapped,
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: const Icon(
                      Icons.local_laundry_service,
                      color: Color(0xFFEF5350),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'ayoeuci Laundry',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 12,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      Text(
                        ' $_username',
                        style: const TextStyle(
                          color: Colors.white70,
                          fontSize: 10,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              Container(
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: IconButton(
                  icon: const Icon(
                    Icons.notifications_outlined,
                    color: Colors.white,
                  ),
                  onPressed: () {},
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(Icons.calendar_today, color: Colors.white, size: 16),
                    const SizedBox(width: 8),
                    Text(
                      _currentDate,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                Row(
                  children: [
                    const SizedBox(width: 8),
                    Text(
                      _currentTime,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(0.2),
              borderRadius: BorderRadius.circular(20),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                const Text(
                  'Pendapatan Hari Ini',
                  style: TextStyle(color: Colors.white70, fontSize: 12),
                ),
                const SizedBox(height: 4),
                const Text(
                  'IDR 10.000.000',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    _buildBalanceChip('Penjualan : 2.000.000'),
                    const SizedBox(width: 8),
                    _buildBalanceChip('Pengeluaran : 8.000.000'),
                    const SizedBox(width: 8),
                    _buildBalanceChip('Transaksi : 120'),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBalanceChip(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.2),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: Colors.black87,
          fontSize: 10,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  Widget _buildMenuGrid() {
    final menuItems = [
      {'image': 'assets/images/pengeluaran.png', 'label': 'Pengeluaran'},
      {'image': 'assets/images/pelanggan.png', 'label': 'Pelanggan'},
      {'image': 'assets/images/karyawan.png', 'label': 'Karyawan'},
      {'image': 'assets/images/transaksi.png', 'label': 'Transaksi'},
      {
        'image': 'assets/images/laporanKeuangan.png',
        'label': 'Laporan Keuangan',
      },
      {'image': 'assets/images/outlet.png', 'label': 'Manage Outlet'},
    ];

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.08),
            blurRadius: 15,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: GridView.builder(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 3,
          crossAxisSpacing: 10,
          mainAxisSpacing: 14,
          childAspectRatio: 1.0,
        ),
        itemCount: menuItems.length,
        itemBuilder: (context, index) {
          return InkWell(
            onTap: () {
              final label = menuItems[index]['label'];
              if (label == 'Pengeluaran') {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const PengeluaranPage()),
                );
              } else if (label == 'Karyawan') {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const KaryawanPage()),
                );
              } else if (label == 'Pelanggan') {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const CustomerPage()),
                );
              }
            },

            borderRadius: BorderRadius.circular(20),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  width: 56,
                  height: 56,
                  decoration: BoxDecoration(
                    color: const Color(0xFFEF5350).withOpacity(0.1),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  padding: const EdgeInsets.all(8),
                  child: Image.asset(
                    menuItems[index]['image'] as String,
                    fit: BoxFit.contain,
                    errorBuilder: (context, error, stackTrace) {
                      return Icon(
                        Icons.image_not_supported,
                        color: Colors.grey[400],
                        size: 36,
                      );
                    },
                  ),
                ),
                const SizedBox(height: 8),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  child: Text(
                    menuItems[index]['label'] as String,
                    textAlign: TextAlign.center,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      fontSize: 10,
                      fontWeight: FontWeight.w500,
                      color: Colors.black87,
                    ),
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildInformasiSection() {
    final infoItems = [
      {
        'gradient': LinearGradient(
          colors: [Color(0xFFEF5350), Color(0xFFFF7043)],
        ),
        'title': 'Promo Spesial',
        'subtitle': 'Diskon hingga 30%',
        'image':
            'https://via.placeholder.com/400x200/EF5350/FFFFFF?text=Promo+Spesial',
      },
      {
        'gradient': LinearGradient(
          colors: [Color(0xFF66BB6A), Color(0xFF26A69A)],
        ),
        'title': 'Info Terbaru',
        'subtitle': 'Layanan Express Tersedia',
        'image':
            'https://via.placeholder.com/400x200/66BB6A/FFFFFF?text=Info+Terbaru',
      },
      {
        'gradient': LinearGradient(
          colors: [Color(0xFF42A5F5), Color(0xFF7E57C2)],
        ),
        'title': 'Pengumuman',
        'subtitle': 'Jadwal Operasional Baru',
        'image':
            'https://via.placeholder.com/400x200/42A5F5/FFFFFF?text=Pengumuman',
      },
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: const [
            Text('⭐', style: TextStyle(fontSize: 16)),
            SizedBox(width: 8),
            Text(
              'Informasi',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ],
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 150,
          child: PageView.builder(
            controller: _infoPageController,
            onPageChanged: (index) {
              setState(() {
                _currentInfoPage = index;
              });
            },
            itemCount: infoItems.length,
            itemBuilder: (context, index) {
              return Padding(
                padding: const EdgeInsets.only(right: 8),
                child: _buildSlideCard(
                  gradient: infoItems[index]['gradient'] as Gradient,
                  title: infoItems[index]['title'] as String,
                  subtitle: infoItems[index]['subtitle'] as String,
                  imageUrl: infoItems[index]['image'] as String,
                ),
              );
            },
          ),
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(
            3,
            (index) => Container(
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: _currentInfoPage == index ? 24 : 8,
              height: 8,
              decoration: BoxDecoration(
                color: _currentInfoPage == index
                    ? Color(0xFFEF5350)
                    : Colors.grey[300],
                borderRadius: BorderRadius.circular(4),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildPromoSection() {
    final promoItems = [
      {
        'gradient': LinearGradient(
          colors: [Color(0xFFEF5350), Color(0xFFFF7043)],
        ),
        'title': 'Penawaran Terbatas',
        'subtitle': 'Cuci + Setrika Hemat',
        'image':
            'https://via.placeholder.com/400x200/EF5350/FFFFFF?text=Penawaran+Terbatas',
      },
      {
        'gradient': LinearGradient(
          colors: [Color(0xFF66BB6A), Color(0xFF26A69A)],
        ),
        'title': 'Member Baru',
        'subtitle': 'Gratis Ongkir Pertama',
        'image':
            'https://via.placeholder.com/400x200/66BB6A/FFFFFF?text=Member+Baru',
      },
      {
        'gradient': LinearGradient(
          colors: [Color(0xFF42A5F5), Color(0xFF7E57C2)],
        ),
        'title': 'Paket Hemat',
        'subtitle': 'Diskon 20% Kiloan',
        'image':
            'https://via.placeholder.com/400x200/42A5F5/FFFFFF?text=Paket+Hemat',
      },
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: const [
            Text('🎁', style: TextStyle(fontSize: 16)),
            SizedBox(width: 8),
            Text(
              'Promo',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
          ],
        ),
        const SizedBox(height: 12),
        SizedBox(
          height: 150,
          child: PageView.builder(
            controller: _promoPageController,
            onPageChanged: (index) {
              setState(() {
                _currentPromoPage = index;
              });
            },
            itemCount: promoItems.length,
            itemBuilder: (context, index) {
              return Padding(
                padding: const EdgeInsets.only(right: 8),
                child: _buildSlideCard(
                  gradient: promoItems[index]['gradient'] as Gradient,
                  title: promoItems[index]['title'] as String,
                  subtitle: promoItems[index]['subtitle'] as String,
                  imageUrl: promoItems[index]['image'] as String,
                ),
              );
            },
          ),
        ),
        const SizedBox(height: 8),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(
            3,
            (index) => Container(
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: _currentPromoPage == index ? 24 : 8,
              height: 8,
              decoration: BoxDecoration(
                color: _currentPromoPage == index
                    ? Color(0xFFEF5350)
                    : Colors.grey[300],
                borderRadius: BorderRadius.circular(4),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSlideCard({
    required Gradient gradient,
    required String title,
    required String subtitle,
    required String imageUrl,
  }) {
    return Container(
      decoration: BoxDecoration(
        gradient: gradient,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Stack(
          children: [
            Positioned.fill(
              child: Image.network(
                imageUrl,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) {
                  return Container(
                    decoration: BoxDecoration(gradient: gradient),
                  );
                },
              ),
            ),
            Positioned.fill(
              child: Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [Colors.black.withOpacity(0.3), Colors.transparent],
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                  ),
                ),
              ),
            ),
            Positioned(
              bottom: 16,
              left: 16,
              right: 16,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: const TextStyle(color: Colors.white, fontSize: 12),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
