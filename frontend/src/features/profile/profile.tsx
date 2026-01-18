import { AppShell } from "@/app/shell";
import { useUser } from "@/app/user-context";
import { useMe } from "@/features/users/queries";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useEffect } from "react";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Mail, User, Shield, Briefcase, ChevronsUpDown } from "lucide-react";
import { useState } from "react";
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

import { ValidSkillsList } from "@/lib/constants";

export function Profile() {
    const { user, setUser } = useUser();
    const { data: me, isLoading } = useMe();
    const [isEditing, setIsEditing] = useState(false);

    // Sync context with backend data
    useEffect(() => {
        if (me) {
            setUser(me);
        }
    }, [me, setUser]);

    if (isLoading && !user) {
        return (
            <AppShell>
                <div className="space-y-4">
                    <Skeleton className="h-8 w-1/3" />
                    <Skeleton className="h-[200px] w-full" />
                </div>
            </AppShell>
        );
    }

    if (!user) {
        return (
            <AppShell>
                <div className="p-4 text-center">User not found. Please log in again.</div>
            </AppShell>
        );
    }

    const userSkills = user.skills || [];

    const toggleSkill = (skill: string) => {
        const newSkills = userSkills.includes(skill)
            ? userSkills.filter((s: string) => s !== skill)
            : [...userSkills, skill];

        setUser({ ...user, skills: newSkills });
        // Note: In a real app, we would also call an API here to persist the change
    };

    return (
        <AppShell>
            <div className="space-y-6">
                <div className="flex justify-between items-end">
                    <div>
                        <h1 className="text-3xl font-bold">Profile</h1>
                        <p className="text-muted-foreground">Your account information</p>
                    </div>
                    <Button variant="outline" onClick={() => setIsEditing(!isEditing)}>
                        {isEditing ? "Done Editing" : "Edit Profile"}
                    </Button>
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

                        {/* Skills */}
                        <div>
                            <label className="text-sm text-muted-foreground flex items-center gap-2 mb-2">
                                <Briefcase className="h-4 w-4" />
                                Skills
                            </label>

                            {isEditing ? (
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="outline" className="w-full md:w-[300px] justify-between">
                                            {userSkills.length > 0
                                                ? `${userSkills.length} skills selected`
                                                : "Select skills..."}
                                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent className="w-[300px]">
                                        {ValidSkillsList.map((skill) => (
                                            <DropdownMenuCheckboxItem
                                                key={skill}
                                                checked={userSkills.includes(skill)}
                                                onCheckedChange={() => toggleSkill(skill)}
                                            >
                                                {skill.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')}
                                            </DropdownMenuCheckboxItem>
                                        ))}
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            ) : (
                                <div className="flex flex-wrap gap-2 mt-1">
                                    {userSkills.length > 0 ? (
                                        userSkills.map(skill => (
                                            <Badge key={skill} variant="outline" className="capitalize">
                                                {skill.replace(/-/g, ' ')}
                                            </Badge>
                                        ))
                                    ) : (
                                        <p className="text-muted-foreground italic">No skills listed</p>
                                    )}
                                </div>
                            )}
                        </div>

                    </CardContent>
                </Card>
            </div>
        </AppShell>
    );
}
