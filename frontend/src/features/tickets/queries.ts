import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Ticket, PaginatedResult, Comment, TicketStats } from '@/lib/types';

export const useTicketStats = () =>
    useQuery({
        queryKey: ['tickets', 'stats'],
        queryFn: () => api<TicketStats>('/api/v1/ticket/stats'),
    });

export const useTickets = (params: { q?: string; state?: string; page?: number }) =>
    useQuery({
        queryKey: ['tickets', params],
        queryFn: () => {
            // Filter out undefined params
            const queryParams = new URLSearchParams();
            Object.entries(params).forEach(([key, value]) => {
                if (value !== undefined && value !== null && value !== '') {
                    queryParams.append(key, String(value));
                }
            });
            return api<PaginatedResult<Ticket>>(`/api/v1/ticket/all?${queryParams.toString()}`);
        }
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
