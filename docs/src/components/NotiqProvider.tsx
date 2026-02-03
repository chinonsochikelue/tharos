import "@collabchron/notiq/styles.css";

export const NotiqProvider = ({ children }: { children: React.ReactNode }) => {
    return <div className="notiq-root">{children}</div>;
};
