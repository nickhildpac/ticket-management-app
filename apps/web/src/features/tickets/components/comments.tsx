import { useComments, useCreateComment } from "../queries";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useState } from "react";
import { Skeleton } from "@/components/ui/skeleton";

export function CommentsSection({ ticketId }: { ticketId: string }) {
    const { data: comments, isLoading } = useComments(ticketId);
    const createComment = useCreateComment();
    const [commentBody, setCommentBody] = useState("");

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (!commentBody.trim()) return;

        createComment.mutate(
            { ticketId, body: commentBody },
            {
                onSuccess: () => {
                    setCommentBody("");
                },
            }
        );
    };

    return (
        <div className="space-y-6">
            <div className="flex flex-col gap-4">
                <h3 className="font-semibold text-lg">Comments</h3>
                <div className="space-y-4">
                    {isLoading ? (
                        <div className="space-y-3">
                            <Skeleton className="h-16 w-full" />
                            <Skeleton className="h-16 w-full" />
                        </div>
                    ) : comments?.length === 0 ? (
                        <p className="text-muted-foreground text-sm">No comments yet.</p>
                    ) : (
                        comments?.map((comment) => (
                            <div key={comment.id} className="flex gap-4 p-4 border rounded-lg bg-card">
                                <Avatar className="h-8 w-8">
                                    <AvatarFallback>{(comment.creator?.first_name || comment.created_by).slice(0, 2).toUpperCase()}</AvatarFallback>
                                </Avatar>
                                <div className="flex-1 space-y-1">
                                    <div className="flex items-center justify-between">
                                        <span className="text-sm font-medium">{comment.creator ? `${comment.creator.first_name} ${comment.creator.last_name}` : comment.created_by}</span>
                                        <span className="text-xs text-muted-foreground">{new Date(comment.created_at).toLocaleString()}</span>
                                    </div>
                                    <p className="text-sm text-foreground/90 whitespace-pre-wrap">{comment.description}</p>
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>

            <div className="space-y-4 pt-4 border-t">
                <h4 className="text-sm font-medium">Add a comment</h4>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <Textarea
                        placeholder="Write your comment..."
                        value={commentBody}
                        onChange={(e) => setCommentBody(e.target.value)}
                        className="min-h-[100px]"
                    />
                    <div className="flex justify-end">
                        <Button type="submit" disabled={createComment.isPending || !commentBody.trim()}>
                            {createComment.isPending ? "Posting..." : "Post Comment"}
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    );
}
