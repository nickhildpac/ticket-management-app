export interface UserInfo {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    role: string;
    skills?: string[];
    created_at: string;
    updated_at?: string;
}
