#!/usr/bin/env bash

URL="http://localhost:8080/users"
N=$1

if [ -z "$N" ]; then
    read -p "🔥 Mau berondong berapa kali hit ke $URL? Masukkan angka (n): " N
fi

if ! [[ "$N" =~ ^[0-9]+$ ]] || [ "$N" -le 0 ]; then
    echo "❌ Error: Masukkan angka bulat positif yang valid (contoh: 10, 50, 100)."
    exit 1
fi

echo ""
echo "🚀 Memulai berondong $N request ke $URL ..."
echo "================================================================"

TMP_RESULT=$(mktemp)
SUCCESS=0
FAIL=0

for (( i=1; i<=N; i++ )); do
    RESPONSE=$(curl -s -w "\n%{http_code}" "$URL")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | sed '$d')

    if [ "$HTTP_CODE" == "200" ]; then
        SERVED_BY=$(echo "$BODY" | grep -o '"served_by": *"[^"]*"' | cut -d'"' -f4)
        PORT=$(echo "$BODY" | grep -o '"server_port": *"[^"]*"' | cut -d'"' -f4)
        printf "Hit #%-4d -> ✅ Status: %s | Server: %-6s (Port: %s)\n" "$i" "$HTTP_CODE" "$SERVED_BY" "$PORT"
        echo "$SERVED_BY (Port $PORT)" >> "$TMP_RESULT"
        ((SUCCESS++))
    else
        printf "Hit #%-4d -> ❌ GAGAL (HTTP Status: %s)\n" "$i" "$HTTP_CODE"
        ((FAIL++))
    fi
done

echo "================================================================"
echo "📊 RINGKASAN ($N HIT):"
echo "   Total Sukses : $SUCCESS"
echo "   Total Gagal  : $FAIL"
echo ""

if [ -s "$TMP_RESULT" ]; then
    echo "📌 Distribusi Beban per Server:"
    sort "$TMP_RESULT" | uniq -c | while read -r count server; do
        printf "   - %-18s : %s kali hit\n" "$server" "$count"
    done
fi

rm -f "$TMP_RESULT"
echo "================================================================"
echo "📝 Cek juga log lengkapnya di: logs/hits.log & logs/nginx-balancer.log"
