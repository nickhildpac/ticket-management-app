export interface User {
    id: string;
    username: string;
    email: string;
    name: string;
    role: 'admin' | 'agent' | 'user';
    createdAt: string;
}

export interface UserResponse {
    user: User;
}
