import { AppShell } from "@/app/shell";
import { useUser } from "@/app/user-context";
import { useMe, useUpdateMySkills } from "@/features/users/queries";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Mail, User, Shield, Briefcase, ChevronsUpDown } from "lucide-react";
import { useCallback, useState } from "react";
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

import { ValidSkillsList } from "@/lib/constants";
import { getAuthUser, setAuthUser } from "@/app/auth";

function skillsEqual(a: string[] | undefined, b: string[]): boolean {
    const x = [...(a ?? [])].sort();
    const y = [...b].sort();
    return x.length === y.length && x.every((v, i) => v === y[i]);
}

export function Profile() {
    const { user } = useUser();
    const { data: me, isLoading } = useMe();
    const updateSkills = useUpdateMySkills();
    const [isEditing, setIsEditing] = useState(false);
    const [draftSkills, setDraftSkills] = useState<string[]>([]);
    const [saveError, setSaveError] = useState<string | null>(null);

    const startEditing = useCallback(() => {
        setSaveError(null);
        setDraftSkills([...(user?.skills ?? [])]);
        setIsEditing(true);
    }, [user?.skills]);

    const finishEditing = useCallback(async () => {
        if (!user) return;
        setSaveError(null);
        const baseline = user.skills ?? [];
        if (!skillsEqual(baseline, draftSkills)) {
            try {
                const updated = await updateSkills.mutateAsync(draftSkills);
                const prev = getAuthUser();
                if (prev) {
                    setAuthUser({
                        ...prev,
                        ...updated,
                        skills: updated.skills ?? [],
                    });
                }
            } catch (e) {
                setSaveError(e instanceof Error ? e.message : "Failed to save skills");
                return;
            }
        }
        setIsEditing(false);
    }, [user, draftSkills, updateSkills]);

    if (isLoading && !user && !me) {
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

    const displaySkills = isEditing ? draftSkills : (user.skills ?? []);

    const toggleSkill = (skill: string) => {
        if (!isEditing) return;
        setDraftSkills((prev) =>
            prev.includes(skill) ? prev.filter((s) => s !== skill) : [...prev, skill]
        );
    };

    return (
        <AppShell>
            <div className="space-y-6">
                <div className="flex justify-between items-end">
                    <div>
                        <h1 className="text-3xl font-bold">Profile</h1>
                        <p className="text-muted-foreground">Your account information</p>
                    </div>
                    <div className="flex flex-col items-end gap-2">
                        {isEditing ? (
                            <div className="flex gap-2">
                                <Button
                                    variant="outline"
                                    onClick={() => {
                                        setSaveError(null);
                                        setIsEditing(false);
                                    }}
                                    disabled={updateSkills.isPending}
                                >
                                    Cancel
                                </Button>
                                <Button
                                    variant="default"
                                    onClick={() => void finishEditing()}
                                    disabled={updateSkills.isPending}
                                >
                                    {updateSkills.isPending ? "Saving…" : "Save"}
                                </Button>
                            </div>
                        ) : (
                            <Button variant="outline" onClick={startEditing}>
                                Edit skills
                            </Button>
                        )}
                        {saveError ? (
                            <p className="text-sm text-destructive max-w-xs text-right">{saveError}</p>
                        ) : null}
                    </div>
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
                                            {displaySkills.length > 0
                                                ? `${displaySkills.length} skills selected`
                                                : "Select skills..."}
                                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent className="w-[300px]">
                                        {ValidSkillsList.map((skill) => (
                                            <DropdownMenuCheckboxItem
                                                key={skill}
                                                checked={displaySkills.includes(skill)}
                                                onCheckedChange={() => toggleSkill(skill)}
                                            >
                                                {skill.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')}
                                            </DropdownMenuCheckboxItem>
                                        ))}
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            ) : (
                                <div className="flex flex-wrap gap-2 mt-1">
                                    {displaySkills.length > 0 ? (
                                        displaySkills.map(skill => (
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
