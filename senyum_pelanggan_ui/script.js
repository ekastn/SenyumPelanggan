let video;
let stream;
let currentPage = 1;
let dataRiwayat = [];
const dataPerHalaman = 5;

function tampilkanHalaman(halaman) {
  document.getElementById("deteksi-section").style.display = "none";
  document.getElementById("riwayat-section").style.display = "none";
  document.getElementById("laporan-section").style.display = "none";

  if (halaman === "deteksi") {
    document.getElementById("deteksi-section").style.display = "block";
    mulaiKamera();
  } else if (halaman === "riwayat") {
    document.getElementById("riwayat-section").style.display = "block";
    loadRiwayat();
  } else if (halaman === "laporan") {
    document.getElementById("laporan-section").style.display = "block";
    loadLaporan();
  }
}

function mulaiKamera() {
  video = document.getElementById("kamera");
  navigator.mediaDevices.getUserMedia({ video: true })
    .then((s) => {
      stream = s;
      video.srcObject = stream;
    })
    .catch((err) => {
      alert("Gagal membuka kamera: " + err.message);
    });
}

async function jalankanDeteksi() {
  const canvas = document.getElementById("snapshot-canvas");
  const ctx = canvas.getContext("2d");

  const frames = [];
  const interval = 500; // 0.5 detik
  const jumlahFrame = 10;

  for (let i = 0; i < jumlahFrame; i++) {
    ctx.drawImage(video, 0, 0, canvas.width, canvas.height);
    const dataURL = canvas.toDataURL("image/jpeg");
    const base64Data = dataURL.split(",")[1];
    frames.push(base64Data);
    await new Promise((res) => setTimeout(res, interval));
  }

  fetch("/api/deteksi", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ frames: frames, interval: 0.5 }),
  })
    .then((res) => res.json())
    .then((data) => {
      alert("Deteksi selesai!");
      tampilkanHalaman("riwayat");
    })
    .catch((err) => alert("Gagal deteksi: " + err.message));
}

function loadRiwayat() {
  fetch("/api/riwayat")
    .then((res) => res.json())
    .then((data) => {
      dataRiwayat = data.reverse();
      currentPage = 1;
      renderRiwayat();
    });
}

function renderRiwayat() {
  const tbody = document.querySelector("#riwayat-table tbody");
  tbody.innerHTML = "";

  const mulai = (currentPage - 1) * dataPerHalaman;
  const akhir = mulai + dataPerHalaman;
  const dataHalaman = dataRiwayat.slice(mulai, akhir);

  dataHalaman.forEach((item) => {
    const tanggal = new Date(item.waktu).toLocaleString("id-ID");
    const tr1 = document.createElement("tr");
    tr1.innerHTML = `
      <td rowspan="3">${tanggal}</td>
      <td>Netral</td>
      <td>${item.presentase.netral}%</td>
      <td>${parseFloat(item.durasi_netral).toFixed(2)} detik</td>
      <td rowspan="3">${item.emosi_dominan}</td>
      <td rowspan="3"><img src="/api${item.path_foto}" width="80" /></td>
    `;
    const tr2 = document.createElement("tr");
    tr2.innerHTML = `
      <td>Bahagia</td>
      <td>${item.presentase.bahagia}%</td>
      <td>${parseFloat(item.durasi_bahagia).toFixed(2)} detik</td>
    `;
    const tr3 = document.createElement("tr");
    tr3.innerHTML = `
      <td>Tidak Bahagia</td>
      <td>${item.presentase.tidak_bahagia}%</td>
      <td>${parseFloat(item.durasi_tidak_bahagia).toFixed(2)} detik</td>
    `;
    tbody.appendChild(tr1);
    tbody.appendChild(tr2);
    tbody.appendChild(tr3);
  });

  document.getElementById("nomor-halaman").textContent = `Halaman ${currentPage}`;
}

function gantiHalaman(arah) {
  const totalHalaman = Math.ceil(dataRiwayat.length / dataPerHalaman);
  currentPage += arah;
  if (currentPage < 1) currentPage = 1;
  if (currentPage > totalHalaman) currentPage = totalHalaman;
  renderRiwayat();
}

function loadLaporan() {
  const filter = document.getElementById("filter-laporan").value;
  const tbody = document.querySelector("#laporan-table tbody");
  tbody.innerHTML = "";

  let url = "/api/riwayat";
  const today = new Date().toISOString().slice(0, 10);
  const bulan = today.slice(0, 7);

  if (filter === "harian") url += `?tanggal=${today}`;
  else if (filter === "mingguan") url += `?minggu=${today}`;
  else if (filter === "bulanan") url += `?bulan=${bulan}`;
  else return;

  fetch(url)
    .then(res => res.json())
    .then(data => {
      if (!data || data.length === 0) return;
      let totalDeteksi = 0;
      let akumulasi = { Netral: 0, Bahagia: 0, "Tidak Bahagia": 0 };

      data.forEach(item => {
        totalDeteksi++;
        akumulasi.Netral += item.durasi_netral;
        akumulasi.Bahagia += item.durasi_bahagia;
        akumulasi["Tidak Bahagia"] += item.durasi_tidak_bahagia;
      });

      const totalWaktu = akumulasi.Netral + akumulasi.Bahagia + akumulasi["Tidak Bahagia"] || 1;
      const dominan = Object.entries(akumulasi).sort((a, b) => b[1] - a[1])[0][0];

      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${filter.toUpperCase()}</td>
        <td>${totalDeteksi}</td>
        <td>
          Netral: ${(akumulasi.Netral / totalWaktu * 100).toFixed(2)}% (${akumulasi.Netral.toFixed(2)}s)<br>
          Bahagia: ${(akumulasi.Bahagia / totalWaktu * 100).toFixed(2)}% (${akumulasi.Bahagia.toFixed(2)}s)<br>
          Tidak Bahagia: ${(akumulasi["Tidak Bahagia"] / totalWaktu * 100).toFixed(2)}% (${akumulasi["Tidak Bahagia"].toFixed(2)}s)
        </td>
        <td>${dominan}</td>
      `;
      tbody.appendChild(tr);
    });
}

function exportToExcel() {
  const filter = document.getElementById("filter-laporan").value;
  const baseUrl = "/api/laporan/export";

  const today = new Date().toISOString().slice(0, 10);
  const thisMonth = today.slice(0, 7);
  const thisYear = today.slice(0, 4);

  let query = "";

  if (filter === "harian") {
    query = `?tanggal=${today}`;
  } else if (filter === "mingguan") {
    query = `?minggu=${today}`;
  } else if (filter === "bulanan") {
    query = `?bulan=${thisMonth}`;
  } else if (filter === "tahunan") {
    query = `?tahun=${thisYear}`;
  } else if (filter === "rentang") {
    const dari = document.getElementById("bulan-dari").value;
    const sampai = document.getElementById("bulan-sampai").value;
    if (!dari || !sampai) {
      alert("Mohon isi bulan dari dan sampai");
      return;
    }
    query = `?dari=${dari}&sampai=${sampai}`;
  }

  // Buka file hasil export di tab baru
  window.open(`${baseUrl}${query}`, "_blank");
}


window.onload = function () {
  tampilkanHalaman("deteksi");
};

// Toggle input bulan jika filter "rentang" dipilih
document.getElementById("filter-laporan").addEventListener("change", function () {
  const rentangFilter = document.getElementById("rentang-bulan-filter");
  if (this.value === "rentang") {
    rentangFilter.style.display = "inline-block";
  } else {
    rentangFilter.style.display = "none";
  }
  loadLaporan(); // panggil ulang laporan saat filter diganti
});

// Trigger otomatis saat bulan rentang diganti
document.getElementById("bulan-dari").addEventListener("change", loadLaporan);
document.getElementById("bulan-sampai").addEventListener("change", loadLaporan);
