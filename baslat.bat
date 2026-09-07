@echo off
chcp 65001 >nul
echo        Database Monitoring Başlatılıyor
echo.
echo Bu dosya projeyi Docker uzerinden tek tikla ayaga kaldirmak icindir.
echo Kodlari indirdiginizde (veya degistirdiginizde) bu dosyayi calistirabilirsiniz.
echo.
echo Konteynerlar baslatiliyor... (docker compose up -d)
docker compose up -d

echo.
echo Proje basariyla baslatildi!
echo Tarayicida http://localhost:3000 adresi aciliyor...
timeout /t 3 >nul
start http://localhost:3000

echo.
echo Cikmak icin bir tusa basin...
pause >nul
