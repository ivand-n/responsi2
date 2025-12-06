package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// --- STRUKTUR DATA ---
type Barang struct {
	id          int
	user_id string
	nama_barang        string
	harga       int
	jumlah      int
	expired string
	created_at	string
}

// Key untuk context (best practice Go agar tidak tabrakan)
type contextKey string

const userIDKey contextKey = "userID"

var db *sql.DB

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(); err != nil {
		log.Println("Peringatan: Tidak menemukan file .env")
	}

	// 2. Koneksi ke Database Supabase
	var err error
	connStr := os.Getenv("DB_URL")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Cek koneksi
	if err = db.Ping(); err != nil {
		log.Fatal("Gagal koneksi ke DB:", err)
	}
	fmt.Println("Sukses terkoneksi ke Supabase!")

	// 3. Routing (Fitur Go 1.22)
	mux := http.NewServeMux()

	// Public Route
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Halo, Server Go Native Market Berjalan!")
	})

	// Protected Routes (Butuh Login)
	// Kita bungkus handler dengan middleware Auth
	mux.Handle("GET /api/barang", authMiddleware(http.HandlerFunc(getBarang)))
	mux.Handle("POST /api/barang", authMiddleware(http.HandlerFunc(createBarang)))
	mux.Handle("PUT /api/barang/{id}", authMiddleware(http.HandlerFunc(updateBarang)))
	mux.Handle("DELETE /api/barang/{id}", authMiddleware(http.HandlerFunc(deleteBarang)))

	// 4. Jalankan Server dengan CORS sederhana
	port := "8080"
	fmt.Println("Server berjalan di port", port)
	http.ListenAndServe(":"+port, enableCORS(mux))
}

// --- MIDDLEWARE AUTHENTICATION ---
// Tugas: Memvalidasi token JWT yang dikirim dari Frontend (hasil login Google)
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		// Format header: "Bearer <token>"
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := []byte(os.Getenv("SUPABASE_JWT_SECRET"))

		// Parse & Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		// Ambil User ID (sub) dari token
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID := claims["sub"].(string) // 'sub' adalah standar UUID user di Supabase
			
			// Masukkan userID ke dalam context request agar bisa dipakai di Handler
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid Token Claims", http.StatusUnauthorized)
		}
	})
}

// --- HANDLERS (CONTROLLERS) ---

func getBarang(w http.ResponseWriter, r *http.Request) {
	// Query Native SQL
	rows, err := db.Query("SELECT * FROM barang")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var result []Barang
	for rows.Next() {
		var b Barang
		// Scan urutan kolom harus sama dengan Query SELECT
		if err := rows.Scan(&b.id, &b.user_id, &b.nama_barang, &b.harga, &b.jumlah, &b.expired, &b.created_at); err != nil {
			log.Println(err)
			continue
		}
		result = append(result, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func createBarang(w http.ResponseWriter, r *http.Request) {
	// Ambil UserID dari Context (hasil middleware)
	userID := r.Context().Value(userIDKey).(string)

	var b Barang
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Insert ke DB
	sqlStatement := `INSERT INTO barang (user_id, nama_barang, harga, jumlah, expired) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := db.QueryRow(sqlStatement, userID, b.nama_barang, b.harga, b.jumlah, b.expired).Scan(&b.id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b.user_id = userID // Set untuk response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
}

func updateBarang(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)
	id := r.PathValue("id") // Fitur Go 1.22 untuk ambil param URL

	var b Barang
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Update hanya jika ID barang cocok DAN UserID pemiliknya cocok (Security)
	sqlStatement := `UPDATE barang SET nama_barang=$1, harga=$2, jumlah=$3, expired=$4 WHERE id=$5 AND user_id=$6`
	res, err := db.Exec(sqlStatement, b.nama_barang, b.harga, b.jumlah, b.expired, id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Barang tidak ditemukan atau anda bukan pemiliknya", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Barang berhasil diupdate"})
}

func deleteBarang(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(string)
	id := r.PathValue("id")

	// Delete hanya jika milik user tersebut
	sqlStatement := `DELETE FROM barang WHERE id=$1 AND user_id=$2`
	res, err := db.Exec(sqlStatement, id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Gagal menghapus (Unauthorized atau ID salah)", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Barang berhasil dihapus"))
}

// --- HELPER CORS ---
// Agar frontend (localhost:3000) bisa akses backend (localhost:8080)
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}