import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Bot, MessageCircle, Send, Sparkles, TrendingUp, Activity } from 'lucide-react';
import { stockService } from '../services/stockService';
import ReactMarkdown from 'react-markdown';

interface SummaryResponse {
  summary: string;
  generated_at: string;
  tokens_used: number;
}

interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  timestamp: Date;
  context?: string; // Store database context for this message
}

interface ConversationMemory {
  summary: string;
  keyTopics: string[];
  lastContext: string;
}

export const AIAssistant = () => {
  const [summary, setSummary] = useState<SummaryResponse | null>(null);
  const [loadingSummary, setLoadingSummary] = useState(false);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [currentMessage, setCurrentMessage] = useState('');
  const [sendingMessage, setSendingMessage] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [conversationMemory, setConversationMemory] = useState<ConversationMemory>({
    summary: '',
    keyTopics: [],
    lastContext: ''
  });

  useEffect(() => {
    loadSummary();
  }, []);

  const loadSummary = async () => {
    setLoadingSummary(true);
    try {
      const response = await fetch('http://localhost:8081/api/stocks/summary');
      if (response.ok) {
        const data = await response.json();
        setSummary(data);
      }
    } catch (error) {
      console.error('Failed to load AI summary:', error);
    } finally {
      setLoadingSummary(false);
    }
  };

  const sendMessage = async () => {
    if (!currentMessage.trim() || sendingMessage) return;

    const userMessage: ChatMessage = {
      role: 'user',
      content: currentMessage,
      timestamp: new Date()
    };

    setChatMessages(prev => [...prev, userMessage]);
    setCurrentMessage('');
    setSendingMessage(true);

    try {
      // Enhanced chat with conversation memory
      const response = await fetch('http://localhost:8081/api/stocks/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          message: currentMessage,
          conversation_memory: conversationMemory,
          recent_messages: chatMessages.slice(-4).map(msg => ({
            role: msg.role,
            content: msg.content
          }))
        })
      });

      if (response.ok) {
        const data = await response.json();
        const assistantMessage: ChatMessage = {
          role: 'assistant',
          content: data.response,
          timestamp: new Date(),
          context: data.context_used
        };
        setChatMessages(prev => [...prev, assistantMessage]);
        
        // Update conversation memory
        if (data.updated_memory) {
          setConversationMemory(data.updated_memory);
        }
      }
    } catch (error) {
      console.error('Failed to send chat message:', error);
      const errorMessage: ChatMessage = {
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: new Date()
      };
      setChatMessages(prev => [...prev, errorMessage]);
    } finally {
      setSendingMessage(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* AI Market Summary */}
      <Card className="glass-card border border-border/50 animate-fade-in">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary/20 animate-glow">
              <TrendingUp className="h-5 w-5 text-primary" />
            </div>
            AI Market Summary
          </CardTitle>
          <CardDescription>
            AI analysis of the 50 most recent analyst ratings.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loadingSummary ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <Bot className="h-4 w-4 animate-spin" />
              Generating AI analysis...
            </div>
          ) : summary ? (
            <div className="space-y-4">
              <p className="text-sm leading-relaxed">{summary.summary}</p>
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Generated: {new Date(summary.generated_at).toLocaleString()}</span>
                <span>Tokens used: {summary.tokens_used}</span>
              </div>
              <Button 
                onClick={loadSummary} 
                variant="outline" 
                size="sm"
                disabled={loadingSummary}
              >
                <Sparkles className="h-4 w-4 mr-2" />
                Refresh Analysis
              </Button>
            </div>
          ) : (
            <p className="text-muted-foreground">Failed to load AI summary</p>
          )}
        </CardContent>
      </Card>

      {/* AI Chat Assistant */}
      <Card className="glass-card border border-border/50 animate-fade-in">
        <CardHeader>
          <CardTitle className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-success/20">
              <Bot className="h-5 w-5 text-success" />
            </div>
            Need Help?
          </CardTitle>
          <CardDescription>
            Ask our AI agent for personalized stock analysis and investment insights
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Dialog open={isOpen} onOpenChange={setIsOpen}>
            <DialogTrigger asChild>
              <Button className="w-full">
                <MessageCircle className="h-4 w-4 mr-2" />
                Chat with AI Assistant
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-2xl max-h-[80vh]">
              <DialogHeader>
                <DialogTitle className="flex items-center gap-2">
                  <Bot className="h-5 w-5 text-primary" />
                  AI Stock Assistant
                </DialogTitle>
                <DialogDescription>
                  Ask questions about stocks, market trends, or get investment advice
                </DialogDescription>
              </DialogHeader>
              
              <div className="flex flex-col h-96">
                {/* Chat Messages */}
                <div className="flex-1 overflow-y-auto space-y-4 p-4 glass-card border border-border/50">
                  {chatMessages.length === 0 ? (
                    <div className="text-center space-y-6">
                      <div className="flex items-center justify-center">
                        <div className="p-4 rounded-xl bg-primary/20 animate-glow">
                          <Bot className="h-12 w-12 text-primary" />
                        </div>
                      </div>
                      <div className="space-y-4">
                        <div>
                          <h3 className="text-lg font-semibold mb-2">AI Stock Assistant</h3>
                          <p className="text-muted-foreground">Ask me anything about stocks, market trends, or get investment advice</p>
                        </div>
                        
                        <div className="glass-card border border-border/50 p-4 text-left space-y-3">
                          <div className="flex items-center gap-2 mb-3">
                            <Sparkles className="h-4 w-4 text-primary" />
                            <span className="font-semibold text-primary">Pro Tips for Better Results</span>
                          </div>
                          
                          <div className="space-y-3 text-sm">
                            <div className="space-y-2">
                              <div className="flex items-start gap-2">
                                <span className="text-destructive font-medium">❌</span>
                                <span className="text-muted-foreground">"What stocks are good?"</span>
                              </div>
                              <div className="flex items-start gap-2">
                                <span className="text-success font-medium">✅</span>
                                <span className="text-foreground font-medium">"Which biotech stocks have recent buy ratings from Goldman Sachs?"</span>
                              </div>
                            </div>
                            
                            <div className="h-px bg-border/50"></div>
                            
                            <div className="space-y-2">
                              <div className="flex items-start gap-2">
                                <span className="text-destructive font-medium">❌</span>
                                <span className="text-muted-foreground">"Tell me about AAPL"</span>
                              </div>
                              <div className="flex items-start gap-2">
                                <span className="text-success font-medium">✅</span>
                                <span className="text-foreground font-medium">"What are AAPL's recent target price changes and analyst ratings?"</span>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  ) : (
                    chatMessages.map((message, index) => {
                      const messageAlignment = message.role === 'user' ? 'justify-end' : 'justify-start';
                      const messageStyle = message.role === 'user' 
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-background border';
                      
                      return (
                        <div
                          key={index}
                          className={`flex ${messageAlignment}`}
                        >
                          <div className={`max-w-[80%] p-3 rounded-lg ${messageStyle} ${message.role === 'assistant' ? 'glass-card' : ''}`}>
                            {message.role === 'assistant' ? (
                              <div className="text-sm prose prose-sm max-w-none dark:prose-invert">
                                <ReactMarkdown 
                                  components={{
                                    p: ({children}) => <p className="mb-2 last:mb-0">{children}</p>,
                                    ol: ({children}) => <ol className="list-decimal list-inside space-y-1 mb-2">{children}</ol>,
                                    ul: ({children}) => <ul className="list-disc list-inside space-y-1 mb-2">{children}</ul>,
                                    li: ({children}) => <li className="text-sm">{children}</li>,
                                    strong: ({children}) => <strong className="font-semibold text-primary">{children}</strong>
                                  }}
                                >
                                  {message.content}
                                </ReactMarkdown>
                              </div>
                            ) : (
                              <p className="text-sm">{message.content}</p>
                            )}
                            <p className="text-xs opacity-70 mt-1">
                              {message.timestamp.toLocaleTimeString()}
                            </p>
                          </div>
                        </div>
                      );
                    })
                  )}
                  {sendingMessage && (
                    <div className="flex justify-start">
                      <div className="glass-card border border-border/50 p-3 rounded-lg">
                        <div className="flex items-center gap-2 text-muted-foreground">
                          <Bot className="h-4 w-4 animate-spin" />
                          <span className="text-sm">Analyzing market data...</span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>

                {/* Chat Input */}
                <div className="space-y-2">
                  {conversationMemory.keyTopics && conversationMemory.keyTopics.length > 0 && (
                    <div className="glass-card border border-border/50 p-2 rounded-lg">
                      <div className="flex items-center gap-2 text-xs">
                        <div className="p-1 rounded bg-primary/20">
                          <Activity className="h-3 w-3 text-primary" />
                        </div>
                        <span className="font-medium text-muted-foreground">Active Context:</span>
                        <span className="text-primary font-medium">{conversationMemory.keyTopics.join(', ')}</span>
                      </div>
                    </div>
                  )}
                  <div className="flex gap-2">
                    <Input
                      value={currentMessage}
                      onChange={(e) => setCurrentMessage(e.target.value)}
                      placeholder="Ask about stocks, trends, or get investment advice..."
                      onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
                      disabled={sendingMessage}
                    />
                    <Button 
                      onClick={sendMessage} 
                      disabled={!currentMessage.trim() || sendingMessage}
                    >
                      <Send className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </DialogContent>
          </Dialog>
        </CardContent>
      </Card>
    </div>
  );
};