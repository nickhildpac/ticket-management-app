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
    state: TicketStateValue | 'open' | 'pending' | 'resolved' | 'closed' | 'cancelled';
    priority: TicketPriorityValue | 'low' | 'medium' | 'high' | 'critical';
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

// Backend returns state as integer: 1=open, 2=pending, 3=resolved, 4=closed, 5=cancelled
export type TicketStateValue = 1 | 2 | 3 | 4 | 5;
export type TicketPriorityValue = 1 | 2 | 3 | 4; // 1=critical, 2=high, 3=medium, 4=low

export const ticketStateMap: Record<TicketStateValue, 'open' | 'pending' | 'resolved' | 'closed' | 'cancelled'> = {
    1: 'open',
    2: 'pending',
    3: 'resolved',
    4: 'closed',
    5: 'cancelled',
};

export const ticketPriorityMap: Record<TicketPriorityValue, 'critical' | 'high' | 'medium' | 'low'> = {
    1: 'critical',
    2: 'high',
    3: 'medium',
    4: 'low',
};

export function getTicketStateString(state: number): 'open' | 'pending' | 'resolved' | 'closed' | 'cancelled' {
    return ticketStateMap[state as TicketStateValue] || 'open';
}

export function getTicketPriorityString(priority: number): 'critical' | 'high' | 'medium' | 'low' {
    return ticketPriorityMap[priority as TicketPriorityValue] || 'medium';
}
