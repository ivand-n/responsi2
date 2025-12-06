import 'package:flutter/material.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:supabase_flutter/supabase_flutter.dart';
import 'package:http/http.dart' as http;

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // 1. Inisialisasi Supabase
  await Supabase.initialize(
    url: 'postgresql://postgres:K9LwXKLcQUye0feI@db.egmwoonrrkywmywoykud.supabase.co:5432/postgres', // Ganti URL Project Supabase
    anonKey: 'sb_publishable_IUunS2pXgmn6lrQZX5K4Pg_Xaa0NPR5', // Ganti Anon Key Supabase
  );

  runApp(const MyApp());
}

final supabase = Supabase.instance.client;

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(home: const AuthGate());
  }
}

// Widget untuk cek apakah user sudah login atau belum
class AuthGate extends StatelessWidget {
  const AuthGate({super.key});

  @override
  Widget build(BuildContext context) {
    return StreamBuilder<AuthState>(
      stream: supabase.auth.onAuthStateChange,
      builder: (context, snapshot) {
        if (snapshot.hasData) {
          final session = snapshot.data!.session;
          // Jika ada sesi (sudah login), masuk ke halaman utama
          if (session != null) {
            return const HomePage();
          }
        }
        // Jika belum, tampilkan halaman login
        return const LoginPage();
      },
    );
  }
}

// --- HALAMAN LOGIN ---
class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  // Fungsi Login Google Native
  Future<void> _googleSignIn() async {
    try {
      // 1. Trigger Google Sign In Native (Android/iOS dialog)
      const webClientId = '10139103392-dbcpi7q4qmoeinuvo6j5idk8u3e0k5hr.apps.googleusercontent.com';
      const iosClientId = '10139103392-dbcpi7q4qmoeinuvo6j5idk8u3e0k5hr.apps.googleusercontent.com';

      final GoogleSignIn googleSignIn = GoogleSignIn(
        clientId: iosClientId,
        serverClientId: webClientId,
      );

      final googleUser = await googleSignIn.signIn();
      final googleAuth = await googleUser!.authentication;

      // 2. Tukar token Google dengan Sesi Supabase
      if (googleAuth.accessToken == null || googleAuth.idToken == null) {
        throw 'No Access Token found.';
      }

      await supabase.auth.signInWithIdToken(
        provider: OAuthProvider.google,
        idToken: googleAuth.idToken!,
        accessToken: googleAuth.accessToken,
      );

      // Jika berhasil, StreamBuilder di AuthGate akan otomatis mengarahkan ke HomePage
    } catch (e) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text('Login Gagal: $e')));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: ElevatedButton.icon(
          onPressed: _googleSignIn,
          icon: const Icon(Icons.login),
          label: const Text("Login dengan Google"),
        ),
      ),
    );
  }
}

// --- HALAMAN UTAMA (KONEKSI KE BACKEND GO) ---
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  String _response = "Belum ada data";

  // Fungsi memanggil Backend Go
  Future<void> _panggilBackendGo() async {
    try {
      // 1. Ambil JWT Token dari sesi Supabase saat ini
      final session = supabase.auth.currentSession;
      if (session == null) return;

      final jwtToken = session.accessToken;

      // 2. Tentukan URL Backend
      // Jika pakai Android Emulator, localhost diganti 10.0.2.2
      // Jika pakai HP Fisik, gunakan IP Laptop (misal 192.168.1.5)
      var url = Uri.parse('http://10.38.110.152:8080/api/barang');

      // 3. Kirim Request dengan Header Authorization
      var response = await http.get(
        url,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer $jwtToken', // <--- INI KUNCINYA
        },
      );

      // 4. Tampilkan Hasil
      setState(() {
        if (response.statusCode == 200) {
          _response = "Sukses Koneksi Go:\n${response.body}";
        } else {
          _response = "Error ${response.statusCode}: ${response.body}";
        }
      });
    } catch (e) {
      setState(() {
        _response = "Gagal koneksi: $e";
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text("Market Go Native"),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () => supabase.auth.signOut(),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            ElevatedButton(
              onPressed: _panggilBackendGo,
              child: const Text("Ambil Data dari Go Backend"),
            ),
            const SizedBox(height: 20),
            Expanded(child: SingleChildScrollView(child: Text(_response))),
          ],
        ),
      ),
    );
  }
}
