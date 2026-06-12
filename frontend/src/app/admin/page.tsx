"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

interface TierForm {
  name: string;
  price: number;
  total_tickets: number;
}

export default function AdminDashboard() {
  const router = useRouter();
  const [eventName, setEventName] = useState("");
  const [tiers, setTiers] = useState<TierForm[]>([
    { name: "", price: 0, total_tickets: 0 },
  ]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState({ text: "", type: "" });

  useEffect(() => {
    const token = localStorage.getItem("token");
    if (!token) {
      alert("Akses ditolak. Silakan login sebagai Admin.");
      router.push("/login");
    }
  }, [router]);

  const addTierRow = () => {
    setTiers([...tiers, { name: "", price: 0, total_tickets: 0 }]);
  };

  const removeTierRow = (index: number) => {
    const newTiers = tiers.filter((_, i) => i !== index);
    setTiers(newTiers);
  };

  const handleTierChange = (index: number, field: keyof TierForm, value: string | number) => {
    const newTiers = [...tiers];
    // @ts-expect-error - TypeScript akan mengeluh karena value bisa string atau number, tapi kita tahu field mana yang diubah
    newTiers[index][field] = field === "name" ? value : Number(value);
    setTiers(newTiers);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setMessage({ text: "", type: "" });

    const token = localStorage.getItem("token");

    const payload = {
      name: eventName,
      ticket_tiers: tiers.map((tier) => ({
        ...tier,
        available_tickets: tier.total_tickets,
      })),
    };

    try {
      const response = await fetch("http://localhost:8080/admin/events", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "Authorization": `Bearer ${token}`,
        },
        body: JSON.stringify(payload),
      });

      const data = await response.json();

      if (response.ok) {
        setMessage({ text: "✅ Event berhasil dibuat dan live di server!", type: "success" });
        setEventName("");
        setTiers([{ name: "", price: 0, total_tickets: 0 }]);
      } else {
        setMessage({ text: `❌ Gagal: ${data.message}`, type: "error" });
      }
    } catch (error) {
      setMessage({ text: "❌ Terjadi kesalahan jaringan.", type: "error" });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 py-10 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        <div className="mb-8 flex justify-between items-center">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Admin Panel</h1>
            <p className="text-gray-500 mt-1">Buat event baru dan kelola kategori tiket.</p>
          </div>
          <button onClick={() => router.push("/")} className="text-blue-600 hover:text-blue-800 font-medium text-sm">
            &larr; Kembali ke Katalog
          </button>
        </div>

        {message.text && (
          <div className={`mb-6 p-4 rounded-lg font-medium ${message.type === "success" ? "bg-green-100 text-green-800" : "bg-red-100 text-red-800"}`}>
            {message.text}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="bg-white px-6 py-8 rounded-xl shadow-sm border border-gray-100">
            <h2 className="text-xl font-semibold text-gray-800 mb-4">Informasi Event</h2>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Nama Konser / Acara</label>
              <input
                type="text"
                required
                value={eventName}
                onChange={(e) => setEventName(e.target.value)}
                placeholder="Contoh: Konser Dewa 19"
                className="w-full border border-gray-300 rounded-lg p-3 focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition-all outline-none"
              />
            </div>
          </div>

          <div className="bg-white px-6 py-8 rounded-xl shadow-sm border border-gray-100">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold text-gray-800">Kategori Tiket</h2>
              <button
                type="button"
                onClick={addTierRow}
                className="text-sm bg-blue-50 text-blue-600 hover:bg-blue-100 px-3 py-1.5 rounded-md font-medium transition-colors"
              >
                + Tambah Kategori
              </button>
            </div>

            <div className="space-y-4">
              {tiers.map((tier, index) => (
                <div key={index} className="flex flex-col sm:flex-row gap-4 items-end bg-gray-50 p-4 rounded-lg border border-gray-200">
                  <div className="flex-1 w-full">
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Nama Kategori</label>
                    <input
                      type="text"
                      required
                      value={tier.name}
                      onChange={(e) => handleTierChange(index, "name", e.target.value)}
                      placeholder="Misal: VIP"
                      className="w-full border border-gray-300 rounded-md p-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                    />
                  </div>
                  <div className="w-full sm:w-1/3">
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Harga (Rp)</label>
                    <input
                      type="number"
                      required
                      min="0"
                      value={tier.price || ""}
                      onChange={(e) => handleTierChange(index, "price", e.target.value)}
                      placeholder="0"
                      className="w-full border border-gray-300 rounded-md p-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                    />
                  </div>
                  <div className="w-full sm:w-1/4">
                    <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1">Kuota</label>
                    <input
                      type="number"
                      required
                      min="1"
                      value={tier.total_tickets || ""}
                      onChange={(e) => handleTierChange(index, "total_tickets", e.target.value)}
                      placeholder="100"
                      className="w-full border border-gray-300 rounded-md p-2 text-sm focus:ring-2 focus:ring-blue-500 outline-none"
                    />
                  </div>
                  {tiers.length > 1 && (
                    <button
                      type="button"
                      onClick={() => removeTierRow(index)}
                      className="text-red-500 hover:bg-red-50 p-2 rounded-md transition-colors"
                      title="Hapus Kategori"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                      </svg>
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>

          <div className="flex justify-end pt-4">
            <button
              type="submit"
              disabled={loading}
              className={`px-8 py-3 rounded-lg text-white font-bold text-lg transition-all ${
                loading ? "bg-gray-400 cursor-not-allowed" : "bg-gray-900 hover:bg-black shadow-lg hover:shadow-xl"
              }`}
            >
              {loading ? "Menyimpan Data..." : "Publish Event"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}