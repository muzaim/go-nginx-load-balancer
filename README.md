# ⚖️ Belajar Load Balancer dengan Go & Nginx (5 Containers)

Proyek ini adalah simulasi pembelajaran **Load Balancing** dan **High Availability (Failover)** menggunakan **Golang**, **Nginx**, dan **Docker Compose**.

Aplikasi Go yang dibuat sangat sederhana (mengembalikan daftar JSON User), lalu di-deploy ke **5 container terpisah** (`8001` s/d `8005`) dengan **Nginx** di depannya sebagai Load Balancer (`8080`). Setiap kali endpoint di-hit, sistem akan mencatat server mana yang melayani request tersebut ke dalam folder [`./logs`](file:///Users/syafrisurya/project/belajar/go-nginx-load-balancer/logs).

---

## 🏗️ Arsitektur Sistem

```
                        ┌──────────────────────┐
   Browser / Client ───►│ Nginx Load Balancer  │
   http://localhost:8080│     (Port: 8080)     │
                        └──────────┬───────────┘
                                   │ Round-Robin + Automatic Failover
         ┌──────────────┬──────────┴───┬──────────────┬──────────────┐
         ▼              ▼              ▼              ▼              ▼
   ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐
   │ go-app-1  │  │ go-app-2  │  │ go-app-3  │  │ go-app-4  │  │ go-app-5  │
   │ Port:8001 │  │ Port:8002 │  │ Port:8003 │  │ Port:8004 │  │ Port:8005 │
   └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘
         └──────────────┴──────────────┼──────────────┴──────────────┘
                                       ▼
                             📂 ./logs/hits.log
                             📂 ./logs/nginx-balancer.log
```

---

## 📁 Struktur Direktori

```text
.
├── Dockerfile              # Multi-stage Dockerfile untuk build aplikasi Go
├── docker-compose.yml      # Orkestrasi 5 container Go (8001-8005) + 1 Nginx (8080)
├── go.mod                  # Go module definition
├── main.go                 # Aplikasi HTTP server Go (JSON User + File Logger)
├── script.sh               # Script otomatis untuk memberondong n kali hit ke port 8080
├── nginx/
│   └── nginx.conf          # Konfigurasi Load Balancer & Health Failover Nginx
└── logs/                   # Folder output log ketika server di-hit (otomatis terisi)
    ├── hits.log            # Log langsung dari aplikasi Go (mencatat server mana yang kena hit)
    └── nginx-balancer.log  # Log dari sisi Nginx (mencatat upstream IP & port)
```

---

## 🚀 Cara Menjalankan Project

### 1. Jalankan semua container
Pastikan Docker Desktop sudah menyala, lalu jalankan perintah berikut di terminal:

```bash
docker compose up -d --build
```

### 2. Cek status container
Pastikan ke-6 container (`go-app-1` s/d `go-app-5` dan `nginx-load-balancer`) berstatus **Up / Running**:

```bash
docker compose ps
```

| Nama Container | Service | Port Host | Peran |
| :--- | :--- | :--- | :--- |
| `nginx-load-balancer` | `nginx` | `http://localhost:8080` | **Pintu Utama (Load Balancer)** |
| `go-app-1` | `app1` | `http://localhost:8001` | Backend Server 1 |
| `go-app-2` | `app2` | `http://localhost:8002` | Backend Server 2 |
| `go-app-3` | `app3` | `http://localhost:8003` | Backend Server 3 |
| `go-app-4` | `app4` | `http://localhost:8004` | Backend Server 4 |
| `go-app-5` | `app5` | `http://localhost:8005` | Backend Server 5 |

---

## 🧪 Skenario Pengujian (Testing)

### Skenario 1: Test Load Balancing dari Browser
1. Buka browser dan akses: **[http://localhost:8080](http://localhost:8080)** (atau `http://localhost:8080/users`).
2. Kamu akan melihat response JSON seperti ini:
   ```json
   {
     "status": "success",
     "served_by": "app-1",
     "server_port": "8001",
     "timestamp": "2026-10-18T06:20:00Z",
     "data": [
       {
         "id": 1,
         "name": "Budi Santoso",
         "email": "budi@example.com",
         "role": "Backend Engineer"
       },
       ...
     ]
   }
   ```
3. Tekan **Refresh (`Cmd + R` / `F5`)** berkali-kali. Perhatikan nilai `"served_by"` dan `"server_port"` yang akan berubah secara berurutan (**Round-Robin**):
   - Hit 1 → `app-1` (`8001`)
   - Hit 2 → `app-2` (`8002`)
   - Hit 3 → `app-3` (`8003`)
   - Hit 4 → `app-4` (`8004`)
   - Hit 5 → `app-5` (`8005`)
   - Hit 6 → Kembali ke `app-1` (`8001`)

---

### Skenario 2: Pantau Log secara Real-Time di Folder Ini
Buka terminal di direktori project ini, lalu pantau file log yang ada di dalam folder `./logs`:

**Melihat log dari sisi Aplikasi Go (`logs/hits.log`):**
```bash
tail -f logs/hits.log
```
Contoh output:
```text
[2026-10-18 06:21:01] HIT -> Server: app-1  (Port: 8001) | Method: GET  | Path: /            | Client: 172.20.0.7:45112
[2026-10-18 06:21:02] HIT -> Server: app-2  (Port: 8002) | Method: GET  | Path: /            | Client: 172.20.0.7:38910
[2026-10-18 06:21:03] HIT -> Server: app-3  (Port: 8003) | Method: GET  | Path: /            | Client: 172.20.0.7:50124
```

**Melihat log dari sisi Nginx Load Balancer (`logs/nginx-balancer.log`):**
```bash
tail -f logs/nginx-balancer.log
```

---

### Skenario 3: Simulasi Server Mati (Failover Test) 🔥
Sekarang kita uji apakah aplikasi tetap berjalan normal ketika salah satu (atau beberapa) container dimatikan.

1. **Matikan salah satu container** (misalnya `go-app-2` di port `8002`):
   ```bash
   docker stop go-app-2
   ```
2. **Refresh browser** di `http://localhost:8080` berkali-kali.
   - Aplikasi **tetap berjalan normal tanpa error** di browser!
   - Di `logs/hits.log`, kamu akan melihat urutan hit otomatis melompati `app-2` (`app-1` → `app-3` → `app-4` → `app-5`).
3. **Coba matikan lebih ekstrem** (misal matikan `go-app-4` juga):
   ```bash
   docker stop go-app-4
   ```
   - Selama minimal masih ada **1 container hidup**, `http://localhost:8080` akan tetap merespons dengan normal.
4. **Nyalakan kembali container yang mati**:
   ```bash
   docker start go-app-2 go-app-4
   ```
   - Setelah menyala kembali, Nginx akan otomatis memasukkan `app-2` dan `app-4` kembali ke dalam rotasi load balancing!

---

### Skenario 4: Berondong Request Otomatis dengan `script.sh` 🚀
Daripada capek refresh browser satu-satu, kamu bisa menggunakan [`script.sh`](file:///Users/syafrisurya/project/belajar/go-nginx-load-balancer/script.sh) untuk memberondong sebanyak `n` kali hit sekaligus dan melihat rekap pembagian bebannya.

**Cara 1 — Interaktif (nanti ditanya mau berapa kali hit):**
```bash
./script.sh
```
Lalu ketik jumlah hit yang kamu inginkan (misal `25`) lalu tekan Enter.

**Cara 2 — Langsung masukkan angka di argumen:**
```bash
./script.sh 25
```

**Contoh Output:**
```text
🚀 Memulai berondong 10 request ke http://localhost:8080/users ...
================================================================
Hit #1    -> ✅ Status: 200 | Server: app-1  (Port: 8001)
Hit #2    -> ✅ Status: 200 | Server: app-2  (Port: 8002)
Hit #3    -> ✅ Status: 200 | Server: app-3  (Port: 8003)
...
================================================================
📊 RINGKASAN HASIL BERONDONG (10 HIT):
   Total Sukses : 10
   Total Gagal  : 0

📌 Distribusi Beban per Server:
   - app-1 (Port 8001)  : 2 kali hit
   - app-2 (Port 8002)  : 2 kali hit
   - app-3 (Port 8003)  : 2 kali hit
   - app-4 (Port 8004)  : 2 kali hit
   - app-5 (Port 8005)  : 2 kali hit
================================================================
```

---

## ⚙️ Penjelasan Cara Kerja di Nginx (`nginx/nginx.conf`)

Bagian inti pengaturan Load Balancer di [`nginx/nginx.conf`](file:///Users/syafrisurya/project/belajar/go-nginx-load-balancer/nginx/nginx.conf):

```nginx
upstream go_backend_cluster {
    server app1:8001 max_fails=2 fail_timeout=5s;
    server app2:8002 max_fails=2 fail_timeout=5s;
    server app3:8003 max_fails=2 fail_timeout=5s;
    server app4:8004 max_fails=2 fail_timeout=5s;
    server app5:8005 max_fails=2 fail_timeout=5s;
}
```

- **`upstream go_backend_cluster`**: Mendefinisikan grup server backend. Secara default menggunakan algoritma **Round-Robin**.
- **`max_fails=2` & `fail_timeout=5s`**: Jika sebuah container gagal dihubungi sebanyak 2 kali, Nginx menandainya sebagai *down* sementara selama 5 detik dan tidak mengirim traffic ke sana.
- **`proxy_next_upstream error timeout ...`**: Jika saat request berjalan ternyata container tujuan mati, Nginx secara transparan langsung mengoper request tersebut ke container berikutnya tanpa menampilkan halaman error ke user di browser.

### Mengganti Algoritma Load Balancing (Opsional)
Kamu bisa bereksperimen mengganti algoritma di blok `upstream` pada [`nginx/nginx.conf`](file:///Users/syafrisurya/project/belajar/go-nginx-load-balancer/nginx/nginx.conf):
1. **Least Connections** (kirim ke server dengan koneksi aktif paling sedikit):
   ```nginx
   upstream go_backend_cluster {
       least_conn;
       server app1:8001;
       ...
   }
   ```
2. **Weighted Round-Robin** (memberi beban lebih besar ke server tertentu):
   ```nginx
   upstream go_backend_cluster {
       server app1:8001 weight=3; # Menerima 3x lipat request dibanding yang lain
       server app2:8002 weight=1;
       ...
   }
   ```
3. **IP Hash** (user dengan IP yang sama selalu diarahkan ke server yang sama):
   ```nginx
   upstream go_backend_cluster {
       ip_hash;
       server app1:8001;
       ...
   }
   ```
*(Setelah mengubah `nginx/nginx.conf`, jalankan `docker compose restart nginx` untuk menerapkan perubahan).*

---

## 🛑 Cara Menghentikan Project

Untuk mematikan dan menghapus semua container:
```bash
docker compose down
```
