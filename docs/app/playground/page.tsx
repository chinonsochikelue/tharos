import { Playground } from '../../components/Playground';

export default function Page() {
    return (
        <div className="flex flex-col min-h-screen bg-fd-background">
            <div className="flex-1">
                <Playground />
            </div>
        </div>
    );
}
