import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { UserInfo } from "@/app/user-types";

export function useMe() {
    return useQuery({
        queryKey: ["me"],
        queryFn: () => api<UserInfo>("/api/v1/me"),
        retry: false,
    });
}

export function useUpdateMySkills() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: (skills: string[]) =>
            api<UserInfo>("/api/v1/me", {
                method: "PATCH",
                body: JSON.stringify({ skills }),
            }),
        onSuccess: (data) => {
            queryClient.setQueryData(["me"], data);
        },
    });
}
