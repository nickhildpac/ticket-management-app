import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Ticket, PaginatedResult, Comment, TicketStats } from '@/lib/types';
import {
    TICKET_LIST_PAGE_SIZE,
    type TicketListSortField,
    type TicketListSortOrder,
} from './queue-state';

export { TICKET_LIST_PAGE_SIZE } from './queue-state';

/** Go ticket-service may return a bare array or { items, total, page, pageSize }. */
export function normalizePaginatedTickets(
    data: PaginatedResult<Ticket> | Ticket[] | undefined,
): PaginatedResult<Ticket> {
    if (!data) {
        return { items: [], total: 0, page: 1, pageSize: TICKET_LIST_PAGE_SIZE };
    }
    if (Array.isArray(data)) {
        return {
            items: data,
            total: data.length,
            page: 1,
            pageSize: data.length || TICKET_LIST_PAGE_SIZE,
        };
    }
    return {
        items: data.items ?? [],
        total: data.total ?? data.items?.length ?? 0,
        page: data.page ?? 1,
        pageSize: data.pageSize ?? TICKET_LIST_PAGE_SIZE,
    };
}

export type TicketListQueryInput = {
    state?: string;
    priority?: string;
    page?: number;
    /** Search box: if purely numeric (optional leading #), sent as ticket_number */
    search?: string;
    limit?: number;
    /** Matches GET ?sort= (default created_at) */
    sortBy?: TicketListSortField;
    /** Matches GET ?order= (default desc) */
    sortOrder?: TicketListSortOrder;
};

function parseTicketNumberFromSearch(raw: string): string | undefined {
    const t = raw.trim().replace(/^#/, '');
    if (t === '') return undefined;
    if (/^\d+$/.test(t)) return t;
    return undefined;
}

/** Query string for GET /api/v1/tickets (see parseListTicketsQuery on the server). */
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

    const sortBy = input.sortBy ?? 'created_at';
    const sortOrder = input.sortOrder ?? 'desc';
    q.set('sort', sortBy);
    q.set('order', sortOrder);

    return q;
}

export const useTicketStats = () =>
    useQuery({
        queryKey: ['tickets', 'stats'],
        queryFn: () => api<TicketStats>('/api/v1/tickets/stats'),
    });

function ticketListQueryKey(scope: 'list' | 'assigned', input: TicketListQueryInput) {
    const page = Math.max(1, input.page ?? 1);
    const state = input.state ?? 'all';
    const priority = input.priority ?? 'all';
    const search = input.search ?? '';
    const sortBy = input.sortBy ?? 'created_at';
    const sortOrder = input.sortOrder ?? 'desc';
    const limit = input.limit ?? TICKET_LIST_PAGE_SIZE;
    return ['tickets', scope, page, state, priority, search, sortBy, sortOrder, limit] as const;
}

export const useTickets = (params: TicketListQueryInput = {}) =>
    useQuery({
        queryKey: ticketListQueryKey('list', params),
        queryFn: async () => {
            const queryParams = buildTicketListQueryParams(params);
            const data = await api<PaginatedResult<Ticket> | Ticket[]>(
                `/api/v1/tickets?${queryParams.toString()}`,
                { cache: 'no-store' }
            );
            return normalizePaginatedTickets(data);
        },
    });

export const useAssignedTickets = (params: TicketListQueryInput = {}) =>
    useQuery({
        queryKey: ticketListQueryKey('assigned', params),
        queryFn: async () => {
            const queryParams = buildTicketListQueryParams(params);
            queryParams.set('assigned_to', 'me');
            const data = await api<PaginatedResult<Ticket> | Ticket[]>(
                `/api/v1/tickets?${queryParams.toString()}`,
                { cache: 'no-store' }
            );
            return normalizePaginatedTickets(data);
        },
    });

export const useTicket = (id: string) =>
    useQuery({
        queryKey: ['ticket', id],
        queryFn: () => api<Ticket>(`/api/v1/tickets/${id}`),
        enabled: !!id,
    });

export const useCreateTicket = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (data: Partial<Ticket>) => api<Ticket>('/api/v1/tickets', { method: 'POST', body: JSON.stringify(data) }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['tickets'] });
        },
    });
};

export const useUpdateTicket = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (p: { id: string; patch: Partial<Ticket> }) => api<Ticket>(`/api/v1/tickets/${p.id}`, { method: 'PATCH', body: JSON.stringify(p.patch) }),
        onSuccess: (_, { id }) => {
            qc.invalidateQueries({ queryKey: ['ticket', id] });
            qc.invalidateQueries({ queryKey: ['tickets'] });
        },
    });
};

export const useComments = (ticketId: string) =>
    useQuery({
        queryKey: ['tickets', ticketId, 'comments'],
        queryFn: () => api<Comment[]>(`/api/v1/tickets/${ticketId}/comments`),
        enabled: !!ticketId,
    });

export const useCreateComment = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (p: { ticketId: string; body: string }) =>
            api<Comment>(`/api/v1/comments`, { method: 'POST', body: JSON.stringify({ ticket_id: p.ticketId, description: p.body }) }),
        onSuccess: (_, { ticketId }) => {
            qc.invalidateQueries({ queryKey: ['tickets', ticketId, 'comments'] });
        },
    });
};
