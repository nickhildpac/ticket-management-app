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

const ticketSchema = z.object({
    title: z.string().min(3, "Title must be at least 3 characters"),
    description: z.string().min(10, "Description must be at least 10 characters"),
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

                                <div className="flex justify-end gap-2">
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
