import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import type { UserInfo } from "@/app/user-context";

export function useMe() {
    return useQuery({
        queryKey: ["me"],
        queryFn: () => api<UserInfo>("/api/v1/me"),
        retry: false,
    });
}
