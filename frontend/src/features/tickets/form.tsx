import { AppShell } from "@/app/shell";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useCreateTicket } from "./queries";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import { ValidSkillsList } from "@/lib/constants";
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuTrigger
} from "@/components/ui/dropdown-menu";
import { Badge } from "@/components/ui/badge";
import { ChevronsUpDown, X } from "lucide-react";

const ticketSchema = z.object({
    title: z.string().min(3, "Title must be at least 3 characters"),
    description: z.string().min(10, "Description must be at least 10 characters"),
    skills: z.array(z.string()),
});

type TicketFormValues = z.infer<typeof ticketSchema>;

export function TicketForm() {
    const navigate = useNavigate();
    const createTicket = useCreateTicket();

    const form = useForm<TicketFormValues>({
        resolver: zodResolver(ticketSchema),
        defaultValues: {
            title: "",
            description: "",
            skills: [],
        },
    });

    const onSubmit = (values: TicketFormValues) => {
        createTicket.mutate(values, {
            onSuccess: () => {
                navigate({ to: '/tickets' });
            }
        });
    };

    return (
        <AppShell>
            <div className="max-w-2xl mx-auto">
                <h1 className="text-2xl font-bold mb-6">Create New Ticket</h1>
                <Card>
                    <CardHeader><CardTitle>Ticket Information</CardTitle></CardHeader>
                    <CardContent>
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                                <FormField
                                    control={form.control}
                                    name="title"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Title</FormLabel>
                                            <FormControl>
                                                <Input placeholder="Brief summary of the issue" {...field} />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="description"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Description</FormLabel>
                                            <FormControl>
                                                <Textarea
                                                    placeholder="Detailed description of the problem..."
                                                    className="min-h-[150px]"
                                                    {...field}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <FormField
                                    control={form.control}
                                    name="skills"
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Required Skills</FormLabel>
                                            <div className="space-y-2">
                                                <div className="flex flex-wrap gap-2 min-h-[40px] p-2 border rounded-md bg-muted/20">
                                                    {field.value?.length > 0 ? (
                                                        field.value.map((skill) => (
                                                            <Badge key={skill} variant="secondary" className="gap-1 pr-1">
                                                                {skill.replace(/-/g, ' ')}
                                                                <button
                                                                    type="button"
                                                                    onClick={() => {
                                                                        field.onChange(field.value.filter((s: string) => s !== skill));
                                                                    }}
                                                                    className="hover:bg-slate-200 rounded-full p-0.5"
                                                                >
                                                                    <X className="h-3 w-3" />
                                                                </button>
                                                            </Badge>
                                                        ))
                                                    ) : (
                                                        <span className="text-sm text-muted-foreground italic p-1">No skills selected</span>
                                                    )}
                                                </div>
                                                <DropdownMenu>
                                                    <DropdownMenuTrigger asChild>
                                                        <Button variant="outline" type="button" className="w-full justify-between">
                                                            Select skills...
                                                            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
                                                        </Button>
                                                    </DropdownMenuTrigger>
                                                    <DropdownMenuContent className="w-[var(--radix-dropdown-menu-trigger-width)]">
                                                        {ValidSkillsList.map((skill) => (
                                                            <DropdownMenuCheckboxItem
                                                                key={skill}
                                                                checked={field.value?.includes(skill)}
                                                                onCheckedChange={(checked) => {
                                                                    if (checked) {
                                                                        field.onChange([...(field.value || []), skill]);
                                                                    } else {
                                                                        field.onChange(field.value?.filter((s: string) => s !== skill));
                                                                    }
                                                                }}
                                                            >
                                                                {skill.split('-').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')}
                                                            </DropdownMenuCheckboxItem>
                                                        ))}
                                                    </DropdownMenuContent>
                                                </DropdownMenu>
                                            </div>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />

                                <div className="flex justify-end gap-2 pt-4">
                                    <Button variant="outline" type="button" onClick={() => navigate({ to: '/tickets' })}>Cancel</Button>
                                    <Button type="submit" disabled={createTicket.isPending}>
                                        {createTicket.isPending ? "Creating..." : "Create Ticket"}
                                    </Button>
                                </div>
                            </form>
                        </Form>
                    </CardContent>
                </Card>
            </div>
        </AppShell>
    );
}
