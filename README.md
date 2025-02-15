flutter create givenchyco_app
cd givenchyco_app
flutter run
import 'package:flutter/material.dart';

void main() {
  runApp(GivenchycoApp());
}

class GivenchycoApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Givenchyco Shop',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: HomeScreen(),
    );
  }
}

class HomeScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Welcome to Givenchyco')),
      body: Center(child: Text('Shop for the best products here!')),
    );
  }
}dependencies:
  flutter:
    sdk: flutter
  http: ^0.13.6
  flutter pub get
  import 'dart:convert';
import 'package:http/http.dart' as http;

class ApiService {
  static const String baseUrl = "https://givenchyco.shop/api"; // Update if needed

  // Fetch products
  static Future<List<dynamic>> getProducts() async {
    final response = await http.get(Uri.parse("$baseUrl/products"));
    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception("Failed to load products");
    }
  }

  // Add item to cart
  static Future<void> addToCart(String productId) async {
    await http.post(Uri.parse("$baseUrl/cart/add"),
        body: jsonEncode({"product_id": productId}),
        headers: {"Content-Type": "application/json"});
  }
}import 'package:flutter/material.dart';
import '../services/api_service.dart';

class ProductScreen extends StatefulWidget {
  @override
  _ProductScreenState createState() => _ProductScreenState();
}

class _ProductScreenState extends State<ProductScreen> {
  List<dynamic> products = [];

  @override
  void initState() {
    super.initState();
    fetchProducts();
  }

  void fetchProducts() async {
    try {
      var productList = await ApiService.getProducts();
      setState(() {
        products = productList;
      });
    } catch (e) {
      print("Error fetching products: $e");
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text("Products")),
      body: products.isEmpty
          ? Center(child: CircularProgressIndicator())
          : ListView.builder(
              itemCount: products.length,
              itemBuilder: (context, index) {
                var product = products[index];
                return ListTile(
                  title: Text(product["name"]),
                  subtitle: Text("\$${product["price"]}"),
                  trailing: ElevatedButton(
                    onPressed: () => ApiService.addToCart(product["id"]),
                    child: Text("Add to Cart"),
                  ),
                );
              },
            ),
    );
  }import 'package:flutter/material.dart';
import 'screens/product_screen.dart';

void main() {
  runApp(GivenchycoApp());
}

class GivenchycoApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Givenchyco Shop',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: HomeScreen(),
    );
  }
}

class HomeScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Welcome to Givenchyco')),
      body: Center(
        child: ElevatedButton(
          onPressed: () {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (context) => ProductScreen()),
            );
          },
          child: Text("Shop Now"),
        ),
      ),
    );
  }
}
}class ApiService {
  static const String baseUrl = "https://givenchyco.shop/api"; // Adjust if needed

  // Fetch cart items
  static Future<List<dynamic>> getCart() async {
    final response = await http.get(Uri.parse("$baseUrl/cart"));
    if (response.statusCode == 200) {
      return jsonDecode(response.body);
    } else {
      throw Exception("Failed to load cart");
    }
  }

  // Add item to cart
  static Future<void> addToCart(String productId) async {
    await http.post(Uri.parse("$baseUrl/cart/add"),
        body: jsonEncode({"product_id": productId}),
        headers: {"Content-Type": "application/json"});
  }

  // Remove item from cart
  static Future<void> removeFromCart(String productId) async {
    await http.post(Uri.parse("$baseUrl/cart/remove"),
        body: jsonEncode({"product_id": productId}),
        headers: {"Content-Type": "application/json"});
  }
}import 'package:flutter/material.dart';
import 'package:flutterwave_standard/flutterwave.dart';

class CheckoutScreen extends StatelessWidget {
  final double amount;

  CheckoutScreen({required this.amount});

  void processPayment(BuildContext context) async {
    final flutterwave = Flutterwave(
      context: context,
      publicKey: "YOUR_FLUTTERWAVE_PUBLIC_KEY", // Replace with your key
      currency: "NGN",
      amount: amount.toString(),
      email: "customer@example.com", // Replace dynamically
      txRef: DateTime.now().millisecondsSinceEpoch.toString(),
      isTestMode: true,
    );

    final response = await flutterwave.charge();
    if (response != null && response.status == "successful") {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("Payment Successful!")),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text("Payment Failed!")),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text("Checkout")),
      body: Center(
        child: ElevatedButton(
          onPressed: () => processPayment(context),
          child: Text("Pay ₦$amount"),
        ),
      ),
    );
  }import 'package:flutter/material.dart';
import 'screens/product_screen.dart';
import 'screens/cart_screen.dart';
import 'screens/checkout_screen.dart';

void main() {
  runApp(GivenchycoApp());
}

class GivenchycoApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Givenchyco Shop',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: HomeScreen(),
      routes: {
        "/cart": (context) => CartScreen(),
        "/checkout": (context) => CheckoutScreen(amount: 1000.0), // Sample
      },
    );
  }
}

class HomeScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Welcome to Givenchyco')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            ElevatedButton(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => ProductScreen()),
                );
              },
              child: Text("Shop Now"),
            ),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: () {
                Navigator.pushNamed(context, "/cart");
              },
              child: Text("View Cart"),
            ),
          ],
        ),
      ),
    );
  }
}import 'package:flutter/material.dart';
import 'screens/product_screen.dart';
import 'screens/cart_screen.dart';
import 'screens/checkout_screen.dart';

void main() {
  runApp(GivenchycoApp());
}

class GivenchycoApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Givenchyco Shop',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: HomeScreen(),
      routes: {
        "/cart": (context) => CartScreen(),
        "/checkout": (context) => CheckoutScreen(amount: 1000.0), // Sample
      },
    );
  }
}

class HomeScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Welcome to Givenchyco')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            ElevatedButton(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => ProductScreen()),
                );
              },
              child: Text("Shop Now"),
            ),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: () {
                Navigator.pushNamed(context, "/cart");
              },
              child: Text("View Cart"),
            ),
          ],
        ),
      ),
    );
  }
}
}import 'package:flutter/material.dart';
import 'screens/product_screen.dart';
import 'screens/cart_screen.dart';
import 'screens/checkout_screen.dart';

void main() {
  runApp(GivenchycoApp());
}

class GivenchycoApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      title: 'Givenchyco Shop',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: HomeScreen(),
      routes: {
        "/cart": (context) => CartScreen(),
        "/checkout": (context) => CheckoutScreen(amount: 1000.0), // Sample
      },
    );
  }
}

class HomeScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Welcome to Givenchyco')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            ElevatedButton(
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (context) => ProductScreen()),
                );
              },
              child: Text("Shop Now"),
            ),
            SizedBox(height: 20),
            ElevatedButton(
              onPressed: () {
                Navigator.pushNamed(context, "/cart");
              },
              child: Text("View Cart"),
            ),
          ],
        ),
      ),
    );
  }
}
