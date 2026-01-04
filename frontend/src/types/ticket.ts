import type { User } from './user';

export type TicketStatus = 'open' | 'in_progress' | 'resolved' | 'closed';
export type TicketPriority = 'low' | 'medium' | 'high' | 'urgent';

export interface Ticket {
    id: string;
    title: string;
    description: string;
    status: TicketStatus;
    priority: TicketPriority;
    assigneeId?: string;
    assignee?: User;
    creatorId: string;
    creator?: User;
    createdAt: string;
    updatedAt: string;
}

export interface Comment {
    id: string;
    content: string;
    ticketId: string;
    authorId: string;
    author?: User;
    createdAt: string;
}
