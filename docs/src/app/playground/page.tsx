import { Playground } from '@/components/Playground';
import { Suspense } from 'react';

export default function Page() {
    return (
        <div className="flex flex-col min-h-screen bg-fd-background">
            <div className="flex-1">
                <Suspense fallback={
                    <div className="h-screen w-screen flex items-center justify-center bg-fd-background">
                        <div className="animate-spin h-8 w-8 border-4 border-orange-600 border-t-transparent rounded-full" />
                    </div>
                }>
                    <Playground />
                </Suspense>
            </div>
        </div>
    );
}
