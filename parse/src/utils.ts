import natural from 'natural';
const { Tokenizer, WordTokenizer } = natural;

export function numberWithSpaces(x: number) {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}


export function cosineSimilarity(str1: string, str2: string) {
    const tokenizer = new WordTokenizer();
    const tokens1 = tokenizer.tokenize(str1.toLowerCase());
    const tokens2 = tokenizer.tokenize(str2.toLowerCase());

    const allTokens = Array.from(new Set([...tokens1, ...tokens2]));
    
    const vector1 = allTokens.map(token => tokens1.filter(t => t === token).length);
    const vector2 = allTokens.map(token => tokens2.filter(t => t === token).length);

    const dotProduct = vector1.reduce((sum, val, i) => sum + val * vector2[i], 0);
    const magnitude1 = Math.sqrt(vector1.reduce((sum, val) => sum + val * val, 0));
    const magnitude2 = Math.sqrt(vector2.reduce((sum, val) => sum + val * val, 0));

    return dotProduct / (magnitude1 * magnitude2);
}
