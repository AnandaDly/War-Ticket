"use client";

import { useState, useEffect } from "react";
import {BuyTicketResponse} from "@/types/ticket";

export default function Home(){
  const [message, setMessage] = useState<string>("");
  const [loading, setLoading] = useState<boolean>(false);
  const [status, setStatus] = useState<"sukses" | "gagal" | "">("");
  const [stock, setStock] = useState<string>("-");

  useEffect(() => {
    const fetchStock = async () => {
      try {
        const response = await fetch("http://localhost:8080/stock");
        const data = await response.json();
        setStock(data.stock);
      } catch (error) {
        console.error("Gagal mengambil stok:", error);
      }
    };

    fetchStock();
  }, []);

  const handleBuyTicket = async ()=>{
    setLoading(true);
    setMessage("");
    
    try{
      const response = await fetch("http://localhost:8080/buy",{
        method:"POST",
      });
      const data:BuyTicketResponse = await response.json();

      setMessage(data.message);
      setStatus(data.status === "sukses" || data.status === "gagal" ? data.status : "");
    }catch(error){
      setMessage("Terjadi kesalahan koneksi ke server.");
      setStatus("gagal");
    }finally{
      setLoading(false);
    }
  };

  return (
    <div className="p-8 max-w-md mx-auto">
      <h1 className="text-2xl font-bold mb-4">Ticket War Dashboard</h1>
      <p className="mb-4 text-gray-700">
      Sisa Tiket Tersedia: <span className="font-bold text-xl">{stock}</span>
      </p>  

      <button
      onClick={handleBuyTicket}
      disabled={loading}
      className={`px-4 py-2 rounded text-white font-semibold transition ${
        loading
        ? "bg-gray-400 cursor-not-allowed"
        : "bg-blue-600 hover:bg-blue-700"
      }`}
      >
        {loading ? "Memproses..." : "Beli Tiket Sekarang"}
      </button>

      {message && (
        <p className={`mt-4 font-medium ${status === "sukses" ? "text-green-600" : "text-red-600"}`}>
          {message}
        </p>
      )}
    </div>
  );
}