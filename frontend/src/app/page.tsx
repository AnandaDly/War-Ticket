"use client";

import { useState, useEffect } from "react";
import { BuyTicketResponse, EventData } from "@/types/ticket";
import { useRouter } from "next/navigation";

export default function Home() {
  const router = useRouter();

  const [eventData, setEventData] = useState<EventData | null>(null);
  const [selectedTier, setSelectedTier] = useState<number | null>(null);
  const [message, setMessage] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [status, setStatus] = useState<"sukses" | "gagal" | "">("");

  useEffect(() => {
    const fetchEvent = async () => {
      try {
        const response = await fetch("http://localhost:8080/events/1");
        if (response.ok) {
          const data = await response.json();
          setEventData(data);
        }
      } catch (error) {
        console.error("Gagal mengambil data event:", error);
      }
    };

    fetchEvent();
  }, []);

  const handleBuyTicket = async () => {
    const token = localStorage.getItem("token");
    if (!token) {
      alert("Anda harus login terlebih dahulu!");
      router.push("/login");
      return;
    }

    if (!selectedTier) {
      alert("Silakan pilih kategori tiket terlebih dahulu!");
      return;
    }

    setLoading(true);
    setMessage("");

    try {
      const response = await fetch("http://localhost:8080/buy", {
        method: "POST",
        headers: {
          "Authorization": `Bearer ${token}`,
          "Content-Type": "application/json", 
        },
        body: JSON.stringify({
          event_id: 1,
          ticket_tier_id: selectedTier,
        }),
      });

      const data: BuyTicketResponse = await response.json();
      setMessage(data.message);
      setStatus(data.status === "sukses" || data.status === "gagal" ? data.status : "");

      if (data.status === "sukses" && eventData) {
        const updatedTiers = eventData.ticket_tiers.map((tier) => {
          if (tier.id === selectedTier) {
            return { ...tier, available_tickets: tier.available_tickets - 1 };
          }
          return tier;
        });
        setEventData({ ...eventData, ticket_tiers: updatedTiers });
      }
    } catch (error) {
      setMessage("Terjadi kesalahan koneksi ke server.");
      setStatus("gagal");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-3xl font-bold mb-6 text-center">
        {eventData ? eventData.name : "Memuat Event..."}
      </h1>

      {/* Tampilan Pilihan Kategori Tiket */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
        {eventData?.ticket_tiers.map((tier) => (
          <div
            key={tier.id}
            onClick={() => setSelectedTier(tier.id)}
            className={`p-4 border rounded-lg cursor-pointer transition-all ${
              selectedTier === tier.id
                ? "border-blue-500 bg-blue-50 shadow-md ring-2 ring-blue-500"
                : "border-gray-200 hover:border-blue-300"
            } ${tier.available_tickets === 0 ? "opacity-50 cursor-not-allowed" : ""}`}
          >
            <h2 className="text-xl font-semibold">{tier.name}</h2>
            <p className="text-gray-600 mb-2">Rp {tier.price.toLocaleString("id-ID")}</p>
            <p className="text-sm font-medium">
              Sisa Tiket:{" "}
              <span className={tier.available_tickets < 10 ? "text-red-500 font-bold" : "text-green-600"}>
                {tier.available_tickets}
              </span>
            </p>
          </div>
        ))}
      </div>

      <button
        onClick={handleBuyTicket}
        disabled={loading || !selectedTier}
        className={`w-full px-4 py-3 rounded text-white font-semibold transition ${
          loading || !selectedTier
            ? "bg-gray-400 cursor-not-allowed"
            : "bg-blue-600 hover:bg-blue-700 shadow-lg"
        }`}
      >
        {loading ? "Memproses..." : "Beli Tiket Sekarang"}
      </button>

      {message && (
        <div className={`mt-4 p-3 rounded text-center font-medium ${status === "sukses" ? "bg-green-100 text-green-700" : "bg-red-100 text-red-700"}`}>
          {message}
        </div>
      )}
    </div>
  );
}