import {
    TICKET_STATE_NUMBER_TO_WIRE,
    type ApiTicketState,
} from "@/lib/ticket-state-contract.generated";

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
    state: TicketStateValue | ApiTicketState;
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

export type TicketStateValue = keyof typeof TICKET_STATE_NUMBER_TO_WIRE extends infer K
    ? K extends `${infer N extends number}`
        ? N
        : never
    : never;
export type TicketPriorityValue = 1 | 2 | 3 | 4; // 1=critical, 2=high, 3=medium, 4=low

export const ticketStateMap = TICKET_STATE_NUMBER_TO_WIRE as unknown as Record<TicketStateValue, ApiTicketState>;

export const ticketPriorityMap: Record<TicketPriorityValue, 'critical' | 'high' | 'medium' | 'low'> = {
    1: 'critical',
    2: 'high',
    3: 'medium',
    4: 'low',
};

export function getTicketStateString(
    state: number
): ApiTicketState {
    return ticketStateMap[state as TicketStateValue] || 'open';
}

export function getTicketPriorityString(priority: number): 'critical' | 'high' | 'medium' | 'low' {
    return ticketPriorityMap[priority as TicketPriorityValue] || 'medium';
}
