import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Ticket, PaginatedResult, Comment, TicketStats } from '@/lib/types';

/** Matches backend defaultTicketListLimit. */
export const TICKET_LIST_PAGE_SIZE = 20;

export type TicketListQueryInput = {
    state?: string;
    priority?: string;
    page?: number;
    /** Search box: if purely numeric (optional leading #), sent as ticket_number */
    search?: string;
    limit?: number;
};

function parseTicketNumberFromSearch(raw: string): string | undefined {
    const t = raw.trim().replace(/^#/, '');
    if (t === '') return undefined;
    if (/^\d+$/.test(t)) return t;
    return undefined;
}

/** Query string for GET /api/v1/ticket/all and /ticket/assigned (see parseListTicketsQuery on the server). */
export function buildTicketListQueryParams(input: TicketListQueryInput): URLSearchParams {
    const limit = input.limit ?? TICKET_LIST_PAGE_SIZE;
    const page = Math.max(1, input.page ?? 1);
    const offset = (page - 1) * limit;

    const q = new URLSearchParams();
    q.set('limit', String(limit));
    q.set('offset', String(offset));

    if (input.state && input.state !== 'all') {
        q.set('state', input.state);
    }
    if (input.priority && input.priority !== 'all') {
        q.set('priority', input.priority);
    }

    const ticketNumber = input.search != null ? parseTicketNumberFromSearch(input.search) : undefined;
    if (ticketNumber) {
        q.set('ticket_number', ticketNumber);
    }

    return q;
}

export const useTicketStats = () =>
    useQuery({
        queryKey: ['tickets', 'stats'],
        queryFn: () => api<TicketStats>('/api/v1/ticket/stats'),
    });

export const useTickets = (params: TicketListQueryInput = {}) =>
    useQuery({
        queryKey: ['tickets', 'list', params],
        queryFn: async () => {
            const queryParams = buildTicketListQueryParams(params);
            const limit = params.limit ?? TICKET_LIST_PAGE_SIZE;
            const page = Math.max(1, params.page ?? 1);
            const response = await api<Ticket[] | PaginatedResult<Ticket>>(
                `/api/v1/ticket/all?${queryParams.toString()}`
            );
            if (Array.isArray(response)) {
                return {
                    items: response,
                    total: response.length,
                    page,
                    pageSize: limit,
                };
            }
            return response;
        },
    });

export const useAssignedTickets = (params: TicketListQueryInput = {}) =>
    useQuery({
        queryKey: ['tickets', 'assigned', params],
        queryFn: async () => {
            const queryParams = buildTicketListQueryParams(params);
            const limit = params.limit ?? TICKET_LIST_PAGE_SIZE;
            const page = Math.max(1, params.page ?? 1);
            const response = await api<Ticket[] | PaginatedResult<Ticket>>(
                `/api/v1/ticket/assigned?${queryParams.toString()}`
            );
            if (Array.isArray(response)) {
                return {
                    items: response,
                    total: response.length,
                    page,
                    pageSize: limit,
                };
            }
            return response;
        },
    });

export const useTicket = (id: string) =>
    useQuery({
        queryKey: ['ticket', id],
        queryFn: () => api<Ticket>(`/api/v1/ticket/${id}`),
        enabled: !!id,
    });

export const useCreateTicket = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: Partial<Ticket>) => api<Ticket>('/api/v1/ticket', { method: 'POST', body: JSON.stringify(data) }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['tickets'] });
        },
    });
};

export const useUpdateTicket = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (p: { id: string; patch: Partial<Ticket> }) => api<Ticket>(`/api/v1/ticket/${p.id}`, { method: 'PATCH', body: JSON.stringify(p.patch) }),
        onSuccess: (_, { id }) => {
            qc.invalidateQueries({ queryKey: ['ticket', id] });
            qc.invalidateQueries({ queryKey: ['tickets'] });
        },
    });
};

export const useComments = (ticketId: string) =>
    useQuery({
        queryKey: ['tickets', ticketId, 'comments'],
        queryFn: () => api<Comment[]>(`/api/v1/ticket/${ticketId}/comments`),
        enabled: !!ticketId,
    });

export const useCreateComment = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (p: { ticketId: string; body: string }) =>
            api<Comment>(`/api/v1/comment`, { method: 'POST', body: JSON.stringify({ ticket_id: p.ticketId, description: p.body }) }),
        onSuccess: (_, { ticketId }) => {
            qc.invalidateQueries({ queryKey: ['tickets', ticketId, 'comments'] });
        },
    });
};
