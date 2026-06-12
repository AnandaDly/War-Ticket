export interface BuyTicketResponse{
    status:string;
    message: string;
    sisa_tiket?: number;
    checkout_url?: string;
}

export interface TicketTier {
  id: number;
  name: string;
  price: number;
  total_tickets: number;
  available_tickets: number;
}

export interface EventData {
  id: number;
  name: string;
  ticket_tiers: TicketTier[];
}