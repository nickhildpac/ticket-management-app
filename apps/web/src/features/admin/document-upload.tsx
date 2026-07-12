import { useRef, useState } from "react";
import { AppShell } from "@/app/shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card";
import { UploadCloud, FileText, CheckCircle2, MinusCircle } from "lucide-react";
import { useIngestDocuments } from "./queries";

export function DocumentUpload() {
    const [files, setFiles] = useState<File[]>([]);
    const inputRef = useRef<HTMLInputElement>(null);
    const ingest = useIngestDocuments();

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (files.length === 0) return;
        ingest.mutate(files, {
            onSuccess: () => {
                setFiles([]);
                if (inputRef.current) inputRef.current.value = "";
            },
        });
    };

    return (
        <AppShell>
            <div className="space-y-8 pb-10">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">Knowledge Base</h1>
                    <p className="text-muted-foreground text-lg">
                        Upload documents to ingest into the AI triage knowledge base.
                    </p>
                </div>

                <Card className="border-none shadow-sm">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <UploadCloud className="h-5 w-5 text-primary" />
                            Upload Documents
                        </CardTitle>
                        <CardDescription>
                            Text documents are chunked, embedded, and made searchable for triage.
                            Binary or empty files are skipped automatically.
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="documents">Files</Label>
                                <Input
                                    id="documents"
                                    ref={inputRef}
                                    type="file"
                                    multiple
                                    onChange={(e) =>
                                        setFiles(e.target.files ? Array.from(e.target.files) : [])
                                    }
                                    disabled={ingest.isPending}
                                />
                                {files.length > 0 && (
                                    <p className="text-sm text-muted-foreground">
                                        {files.length} file{files.length === 1 ? "" : "s"} selected
                                    </p>
                                )}
                            </div>

                            <Button type="submit" disabled={ingest.isPending || files.length === 0}>
                                {ingest.isPending ? "Ingesting..." : "Ingest documents"}
                            </Button>

                            {ingest.isError && (
                                <p className="text-sm text-red-500">
                                    {(ingest.error as Error).message}
                                </p>
                            )}
                        </form>

                        {ingest.isSuccess && (
                            <div className="mt-6 space-y-3">
                                <p className="text-sm font-medium">
                                    Ingested {ingest.data.total_chunks} chunk
                                    {ingest.data.total_chunks === 1 ? "" : "s"} from{" "}
                                    {ingest.data.files.length} file
                                    {ingest.data.files.length === 1 ? "" : "s"}.
                                </p>
                                <ul className="rounded-lg border divide-y">
                                    {ingest.data.files.map((f) => (
                                        <li
                                            key={f.source}
                                            className="flex items-center justify-between gap-3 px-4 py-2 text-sm"
                                        >
                                            <span className="flex items-center gap-2 truncate">
                                                <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                                                <span className="truncate">{f.source}</span>
                                            </span>
                                            {f.skipped ? (
                                                <span className="flex items-center gap-1 text-muted-foreground shrink-0">
                                                    <MinusCircle className="h-4 w-4" />
                                                    skipped{f.reason ? ` (${f.reason})` : ""}
                                                </span>
                                            ) : (
                                                <span className="flex items-center gap-1 text-green-600 shrink-0">
                                                    <CheckCircle2 className="h-4 w-4" />
                                                    {f.chunks} chunk{f.chunks === 1 ? "" : "s"}
                                                </span>
                                            )}
                                        </li>
                                    ))}
                                </ul>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </AppShell>
    );
}
