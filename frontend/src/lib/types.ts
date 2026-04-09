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
    ticket_number: number;
    title: string;
    description: string;
    state: TicketStateValue | 'open' | 'pending' | 'in progress' | 'resolved' | 'closed' | 'cancelled';
    priority: TicketPriorityValue | 'low' | 'medium' | 'high' | 'critical';
    assigned_to: string[];
    created_by: string;
    creator?: UserInfo;
    category?: string;
    skills: string[];
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

// Backend domain (internal/domain/ticket.go): 1=open, 2=pending, 3=in progress, 4=resolved, 5=closed, 6=cancelled
export type TicketStateValue = 1 | 2 | 3 | 4 | 5 | 6;
export type TicketPriorityValue = 1 | 2 | 3 | 4; // 1=critical, 2=high, 3=medium, 4=low

export const ticketStateMap: Record<
    TicketStateValue,
    'open' | 'pending' | 'in progress' | 'resolved' | 'closed' | 'cancelled'
> = {
    1: 'open',
    2: 'pending',
    3: 'in progress',
    4: 'resolved',
    5: 'closed',
    6: 'cancelled',
};

export const ticketPriorityMap: Record<TicketPriorityValue, 'critical' | 'high' | 'medium' | 'low'> = {
    1: 'critical',
    2: 'high',
    3: 'medium',
    4: 'low',
};

export function getTicketStateString(
    state: number
): 'open' | 'pending' | 'in progress' | 'resolved' | 'closed' | 'cancelled' {
    return ticketStateMap[state as TicketStateValue] || 'open';
}

export function getTicketPriorityString(priority: number): 'critical' | 'high' | 'medium' | 'low' {
    return ticketPriorityMap[priority as TicketPriorityValue] || 'medium';
}
