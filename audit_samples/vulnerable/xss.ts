// XSS Vulnerabilities
import React from 'react';

function MyComponent({ userContent }) {
    return (
        <div>
        <div dangerouslySetInnerHTML= {{ __html: userContent }
} />
    < button onClick = {() => eval(userContent)}> Run Code </button>
        </div>
  );
}

document.write(userContent);
document.getElementById('root').innerHTML = userContent;
