import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { User } from '@/lib/types';

// Get basic user list for assignment (available to all authenticated users)
export const useUsersForAssignment = () =>
    useQuery({
        queryKey: ['users'],
        queryFn: () => api<User[]>('/api/v1/users'),
    });

export const useUsers = () =>
    useQuery({
        queryKey: ['admin', 'users'],
        queryFn: () => api<User[]>('/api/v1/admin/users'),
    });

export const useUpdateUserRole = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (p: { id: string; role: string }) =>
            api<User>(`/api/v1/admin/users/${p.id}/role`, {
                method: 'PUT',
                body: JSON.stringify({ role: p.role }),
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['admin', 'users'] });
        },
    });
};

export const useDeleteUser = () => {
    const qc = useQueryClient();
    return useMutation({
        mutationFn: (id: string) =>
            api(`/api/v1/admin/users/${id}`, {
                method: 'DELETE',
            }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ['admin', 'users'] });
        },
    });
};

// Mirrors the ai-service IngestResponse (proxied via the Go /api/v1/admin/documents endpoint).
export type IngestedFile = {
    source: string;
    chunks: number;
    skipped: boolean;
    reason: string | null;
};

export type IngestResponse = {
    files: IngestedFile[];
    total_chunks: number;
};

// Upload documents to the knowledge base. Sends multipart FormData (field "files"),
// which the api() helper forwards without a JSON Content-Type override.
export const useIngestDocuments = () =>
    useMutation({
        mutationFn: (files: File[]) => {
            const form = new FormData();
            for (const file of files) {
                form.append('files', file);
            }
            return api<IngestResponse>('/api/v1/admin/documents', {
                method: 'POST',
                body: form,
            });
        },
    });
