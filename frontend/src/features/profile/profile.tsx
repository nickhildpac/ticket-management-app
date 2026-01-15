import { AppShell } from "@/app/shell";
import { useUser } from "@/app/user-context";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Mail, User, Shield } from "lucide-react";

export function Profile() {
    const { user } = useUser();

    if (!user) {
        return (
            <AppShell>
                <div className="space-y-4">
                    <Skeleton className="h-8 w-1/3" />
                    <Skeleton className="h-[200px] w-full" />
                </div>
            </AppShell>
        );
    }

    return (
        <AppShell>
            <div className="space-y-6">
                <div>
                    <h1 className="text-3xl font-bold">Profile</h1>
                    <p className="text-muted-foreground">Your account information</p>
                </div>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <User className="h-5 w-5" />
                            Personal Information
                        </CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-6">
                        {/* Name */}
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div>
                                <label className="text-sm text-muted-foreground">First Name</label>
                                <p className="text-lg font-medium">{user.first_name}</p>
                            </div>
                            <div>
                                <label className="text-sm text-muted-foreground">Last Name</label>
                                <p className="text-lg font-medium">{user.last_name}</p>
                            </div>
                        </div>

                        {/* Email */}
                        <div>
                            <label className="text-sm text-muted-foreground">Email Address</label>
                            <div className="flex items-center gap-2 mt-1">
                                <Mail className="h-4 w-4 text-muted-foreground" />
                                <p className="text-lg font-medium">{user.email}</p>
                            </div>
                        </div>

                        {/* Role */}
                        <div>
                            <label className="text-sm text-muted-foreground">Role</label>
                            <div className="flex items-center gap-2 mt-1">
                                <Shield className="h-4 w-4 text-muted-foreground" />
                                <Badge variant={user.role === 'admin' ? 'default' : 'secondary'} className="capitalize">
                                    {user.role}
                                </Badge>
                            </div>
                        </div>

                    </CardContent>
                </Card>
            </div>
        </AppShell>
    );
}
