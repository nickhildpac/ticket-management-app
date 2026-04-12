import { QueryClient } from "@tanstack/react-query";

/** App-wide client so auth helpers can align the `me` query with the session. */
export const queryClient = new QueryClient();
