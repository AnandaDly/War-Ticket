"use client";

import { useState, useEffect } from "react";
import { EventData } from "@/types/ticket";
import { useRouter } from "next/navigation";

export default function Catalog() {
  const [events, setEvents] = useState<EventData[]>([]);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const fetchAllEvents = async () => {
      try {
        const response = await fetch("http://localhost:8080/events");
        if (response.ok) {
          const data = await response.json();
          setEvents(data);
        }
      } catch (error) {
        console.error("Gagal memuat katalog:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchAllEvents();
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-5xl mx-auto">
        <div className="flex justify-between items-end mb-10">
          <div>
            <h1 className="text-4xl font-extrabold text-gray-900 tracking-tight">Katalog Event</h1>
            <p className="text-gray-500 mt-2 text-lg">Temukan konser dan acara terbaik untukmu.</p>
          </div>
          <div className="space-x-4">
            <button onClick={() => router.push("/login")} className="text-blue-600 font-medium hover:underline">Login</button>
            <button onClick={() => router.push("/admin")} className="text-gray-600 font-medium hover:underline">Admin</button>
          </div>
        </div>

        {loading ? (
          <div className="flex justify-center py-20">
            <div className="animate-pulse text-xl text-gray-400 font-medium">Memuat event menarik...</div>
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-8">
            {events.map((event) => {
              const lowestPrice = Math.min(...event.ticket_tiers.map(t => t.price));
              
              return (
                <div 
                  key={event.id} 
                  onClick={() => router.push(`/events/${event.id}`)}
                  className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden cursor-pointer hover:shadow-xl hover:-translate-y-1 transition-all duration-300 flex flex-col"
                >
                  <div className="h-48 bg-gradient-to-br from-blue-500 to-indigo-600 w-full relative">
                    <div className="absolute inset-0 bg-black bg-opacity-20 flex items-center justify-center">
                      <span className="text-white font-bold text-2xl opacity-50">TICKET WAR</span>
                    </div>
                  </div>
                  
                  <div className="p-6 flex-1 flex flex-col">
                    <h2 className="text-2xl font-bold text-gray-900 mb-2">{event.name}</h2>
                    <p className="text-sm text-gray-500 mb-4">{event.ticket_tiers.length} Kategori Tiket</p>
                    
                    <div className="mt-auto pt-4 border-t border-gray-100 flex justify-between items-center">
                      <div>
                        <p className="text-xs text-gray-400 uppercase font-semibold">Mulai dari</p>
                        <p className="text-lg font-bold text-blue-600">Rp {lowestPrice.toLocaleString("id-ID")}</p>
                      </div>
                      <span className="bg-blue-50 text-blue-600 text-sm font-bold px-3 py-1 rounded-full">
                        Beli &rarr;
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}