# Sokoban Fruit Fly

This project implements a single Sokoban agent whose observations are encoded by a
frozen Fruit-Fly-inspired spiking network and whose four-action linear policy is
trained with on-policy REINFORCE and an internal value baseline.

The lightweight network is an engineering simulation inspired by fruit-fly neural
processing; it is not a biological connectome. It uses seeded random synaptic weights
and leaky integrate-and-fire neurons. The optional connectome mode reads the original
repository's `FLYGRAPH2` graph and `meta.json`; its wiring and weights are frozen. In
both modes only the policy and its baseline learn.

## Run

```text
go test ./...
go vet ./...
go run ./cmd/fly games
go run ./cmd/fly download -raw-dir data/raw
go run ./cmd/fly features -level level01
go run ./cmd/fly train -episodes 100 -output-dir results
go run ./cmd/fly eval -output-dir results
go run ./cmd/fly play -output-dir results
```

The connectome dataset can be downloaded automatically. This downloads the same three
MaleCNS v1.0 Feather tables used by the original repository (about 1.1 GB):

```text
go run ./cmd/fly download -raw-dir data/raw
```

The downloader validates the remote size, writes to a temporary file, and only
publishes a completed file. Existing files with the expected size are skipped. It
does not run implicitly during training.

To use the same MaleCNS connectome graph format as the reference repository, build
`graph.bin` and `meta.json` from those tables with the original repository's `fly build`
(or its equivalent build step), then select the connectome explicitly:

```text
go run ./cmd/fly download -raw-dir ..\fruit-fly\data\raw
go run ..\fruit-fly\cmd\fly build
go run ./cmd/fly features -network-mode connectome -graph-dir ..\fruit-fly\data\processed
go run ./cmd/fly train -network-mode connectome -graph-dir ..\fruit-fly\data\processed
```

Training writes `training.csv`, `training.json`, and `checkpoint.json`. The command
line currently uses the built-in levels in `internal/games/sokoban/levels.go`; the
matching text copies are in `data/levels`.

## Reinforcement Learning

Project ini menggunakan **on-policy REINFORCE** untuk melatih policy linear pada
lingkungan Sokoban. Jaringan retina dan spiking network (baik mode `lightweight`
maupun `connectome`) bersifat frozen. Parameter yang dipelajari hanya bobot dan bias
policy empat aksi serta linear value baseline.

### Aturan dan Reward

Satu episode dimulai dari posisi awal level dan berakhir ketika level selesai,
terjadi deadlock, tidak ada aksi legal, atau jumlah langkah mencapai `-max-steps`.
Pada setiap langkah:

- Aksi yang menabrak dinding atau mendorong box ke posisi yang tidak valid dianggap
  tidak legal dan dikeluarkan dari kandidat policy.
- Langkah yang legal tetapi tidak memindahkan pemain mendapat reward `-2`.
- Langkah legal yang memindahkan pemain mendapat reward `-1`.
- Menyelesaikan level memberikan reward tambahan `+100`.
- Deadlock memberikan penalti tambahan `-10`.
- Episode yang mencapai batas langkah tanpa selesai ditandai sebagai timeout.

Dengan demikian, agent didorong untuk menyelesaikan level dengan langkah sesedikit
mungkin dan menghindari deadlock. Reward yang digunakan adalah reward episode yang
dihasilkan oleh implementasi saat ini; tidak ada reward shaping tambahan dari file
konfigurasi.

### Proses Pengambilan Keputusan

Untuk setiap state, proses pengambilan keputusan adalah:

1. State Sokoban diubah menjadi empat channel input: wall, target, box, dan player.
2. Retina meneruskan input ke frozen spiking network. Aktivitas spike yang
   dihasilkan menjadi feature vector berdimensi tetap.
3. Policy menghitung logit untuk empat aksi dengan `z = W x + b`.
4. Logit aksi ilegal dimask, lalu logit aksi legal diubah menjadi probabilitas
   menggunakan masked softmax yang stabil secara numerik.
5. Saat training, satu aksi dipilih secara stochastic berdasarkan probabilitas
   policy. Saat `eval` dan `play`, aksi legal dengan probabilitas terbesar dipilih
   secara greedy.

### Proses Pembelajaran

Pembelajaran dilakukan secara bertahap melalui sejumlah episode. Pada awal setiap
episode, lingkungan mengembalikan agent ke kondisi awal level. Agent kemudian
berinteraksi dengan papan: membaca kondisi saat ini, memilih aksi, menerima reward,
dan berpindah ke kondisi berikutnya. Selama fase ini, pemilihan aksi bersifat
stochastic, sehingga aksi dengan probabilitas lebih besar lebih sering dipilih,
tetapi aksi legal lainnya tetap dapat dicoba. Mekanisme ini memberi kesempatan agent
untuk mengeksplorasi strategi yang berbeda.

Setiap episode menyimpan urutan kondisi, aksi, dan reward yang dialami agent. Setelah
episode berakhir, seluruh reward digunakan untuk menilai kualitas keputusan-keputusan
sebelumnya. Reward masa depan diberi bobot yang semakin kecil berdasarkan discount
factor. Nilai default discount factor adalah `0.99`, sehingga hasil jangka panjang,
termasuk keberhasilan menyelesaikan level, tetap berpengaruh kuat terhadap pembaruan
policy.

Project ini menggunakan value baseline untuk memperkirakan hasil yang diharapkan dari
setiap kondisi. Perbedaan antara hasil aktual dan perkiraan tersebut disebut
*advantage*. Jika advantage positif, keputusan yang diambil dianggap lebih baik dari
perkiraan dan probabilitasnya dinaikkan. Jika advantage negatif, probabilitas
keputusan tersebut diturunkan. Baseline juga diperbarui agar semakin akurat dalam
memperkirakan hasil episode berikutnya.

Pembaruan parameter dilakukan setelah satu episode selesai, bukan setelah setiap
langkah. Learning rate mengatur besar perubahan policy dan baseline pada setiap
pembaruan. Jaringan retina dan jaringan spiking tetap beku; proses pembelajaran hanya
mengubah policy empat aksi dan value baseline. Seed yang sama membuat inisialisasi
jaringan ringan dan pengambilan keputusan stochastic dapat diulang secara konsisten.

### Proses Evaluasi

Evaluasi digunakan untuk mengukur kemampuan policy yang sudah dilatih tanpa mengubah
parameternya. Sistem memuat checkpoint hasil training, menggunakan level dan
konfigurasi jaringan yang dipilih, lalu menjalankan episode dari kondisi awal.
Berbeda dari training, evaluasi tidak melakukan eksplorasi dan tidak memperbarui
policy. Pada setiap kondisi, agent memilih aksi legal dengan probabilitas tertinggi,
sehingga hasil evaluasi menunjukkan perilaku terbaik yang dipilih oleh checkpoint.

Selama evaluasi, sistem mencatat apakah level berhasil diselesaikan, total reward,
jumlah langkah, dan waktu inferensi. Reward membantu menunjukkan kualitas keseluruhan
perjalanan agent, jumlah langkah menunjukkan efisiensi penyelesaian, sedangkan waktu
inferensi menggambarkan biaya komputasi pengambilan keputusan. Jika agent mengalami
deadlock, tidak memiliki aksi legal, atau mencapai batas langkah, episode dinyatakan
tidak berhasil.

Perbandingan antarhasil evaluasi hanya bermakna jika level, seed, jumlah neuron,
network mode, dan checkpoint digunakan secara konsisten. Dengan cara ini, perubahan
hasil dapat dikaitkan dengan policy yang dilatih, bukan dengan perubahan konfigurasi
lingkungan atau encoder.

## Architecture

`Sokoban state -> retina encoder -> frozen LIF network -> activity vector -> masked
linear policy -> action`. The environment validates box/target counts, supports
`# . $ @ * +`, detects corner deadlocks, and enforces a configurable episode limit.

For a transition, REINFORCE uses `A_t = G_t - V(s_t)`. The policy gradient is
`-A_t * grad(log pi(a_t|s_t))`; the value head is trained toward the discounted
return. Illegal actions are excluded before softmax.

## Limitations

The current version provides a lightweight network by default and small deterministic
levels. Connectome mode requires the downloaded tables and a built `FLYGRAPH2` graph.
Training metrics are execution outputs and are not pre-filled with experimental claims.
