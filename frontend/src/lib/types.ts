export type Role = 'user' | 'agent' | 'admin';

export interface User {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    role: Role;
    created_at: string;
}

export interface UserInfo {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
}

export interface Ticket {
    id: string;
    title: string;
    description: string;
    state: 'open' | 'pending' | 'resolved' | 'closed' | 'cancelled';
    priority: 'low' | 'medium' | 'high' | 'critical';
    assigned_to: string[];
    created_by: string;
    creator?: UserInfo;
    category?: string;
    created_at: string;
    updated_at: string;
}

export interface TicketStats {
    total: number;
    open: number;
    pending: number;
    resolved: number;
    mine: number;
}

export interface Comment {
    id: string;
    ticket_id: string;
    created_by: string;
    creator?: UserInfo;
    description: string;
    created_at: string;
}

export interface PaginatedResult<T> {
    items: T[];
    total: number;
    page: number;
    pageSize: number;
}
