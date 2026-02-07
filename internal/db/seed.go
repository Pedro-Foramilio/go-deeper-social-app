package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"

	"github.com/Pedro-Foramilio/social/internal/store"
)

var (
	usernames = []string{
		"john_doe", "jane_smith", "bob_jones", "alice_wonder", "charlie_brown",
		"david_clark", "emma_watson", "frank_miller", "grace_hopper", "henry_ford",
		"mike_johnson", "sarah_wilson", "tom_anderson", "lisa_taylor", "kevin_white",
		"maria_garcia", "james_martin", "anna_lee", "robert_hall", "sophia_king",
		"daniel_wright", "olivia_scott", "william_green", "emily_adams", "joseph_baker",
	}

	passwords = []string{
		"password123", "securepass", "mypassword", "pass1234", "qwerty123",
		"welcome123", "admin123", "user1234", "test1234", "sample123",
		"secure456", "mypass789", "password456", "welcome456", "admin456",
		"user5678", "test5678", "sample456", "password789", "secure789",
		"mypass321", "qwerty456", "welcome789", "admin789", "user9012",
	}

	titles = []string{
		"Getting Started with Go", "Understanding Concurrency", "Best Practices in Web Development",
		"Introduction to Microservices", "Database Design Patterns", "API Development Guide",
		"Mastering Docker Containers", "Cloud Computing Basics", "Security Best Practices",
		"Testing Strategies", "Code Review Tips", "Debugging Techniques",
		"Performance Optimization", "Scalability Patterns", "DevOps Fundamentals",
		"Continuous Integration", "Monitoring and Logging", "Error Handling",
		"Design Patterns", "Clean Code Principles", "Agile Methodologies",
		"Git Workflows", "REST API Design", "GraphQL Introduction",
		"Authentication and Authorization",
	}

	contents = []string{
		"This is a comprehensive guide to getting started with programming.",
		"Learn how to write efficient and maintainable code.",
		"Explore advanced concepts and real-world applications.",
		"A deep dive into modern software architecture.",
		"Best practices for building scalable applications.",
		"Tips and tricks from industry experts.",
		"Understanding the fundamentals of system design.",
		"How to improve your development workflow.",
		"Common pitfalls and how to avoid them.",
		"Practical examples and use cases.",
		"Step-by-step tutorial for beginners.",
		"Advanced techniques for experienced developers.",
		"Industry standards and conventions.",
		"Tools and frameworks you should know.",
		"Real-world case studies and examples.",
		"Performance benchmarks and optimization tips.",
		"Security considerations and implementations.",
		"Testing methodologies and frameworks.",
		"Documentation best practices.",
		"Community resources and learning materials.",
		"Troubleshooting common issues.",
		"Integration with third-party services.",
		"Migration strategies and approaches.",
		"Code organization and structure.",
		"Deployment and release management.",
	}

	tags = []string{
		"golang", "programming", "tutorial", "webdev", "backend",
		"frontend", "database", "devops", "cloud", "docker",
		"kubernetes", "microservices", "api", "rest", "graphql",
		"security", "testing", "performance", "scalability", "architecture",
		"design-patterns", "best-practices", "tips", "guide", "learning",
	}

	comments = []string{
		"Great article! Really helped me understand the concept.",
		"Thanks for sharing this valuable information.",
		"This is exactly what I was looking for!",
		"Very well explained. Keep up the good work!",
		"Interesting perspective on this topic.",
		"Could you elaborate more on this point?",
		"I disagree with some of your conclusions here.",
		"Excellent tutorial, very easy to follow.",
		"This saved me hours of debugging. Thank you!",
		"Looking forward to more content like this.",
		"Brilliant explanation! Very clear and concise.",
		"I have a question about the implementation details.",
		"This approach seems overly complicated to me.",
		"Fantastic resource for beginners!",
		"Would love to see a video tutorial on this.",
		"I tried this but ran into some issues.",
		"Best explanation I've found on this topic so far.",
		"Not sure I agree with this methodology.",
		"Very comprehensive guide. Appreciate the effort!",
		"Can you provide more examples?",
		"This needs to be updated for the latest version.",
		"Simple yet effective approach. Love it!",
		"I've been struggling with this for weeks. Finally solved!",
		"Great work! Looking forward to part 2.",
		"The code examples are really helpful.",
		"I think there's a typo in the third paragraph.",
		"This is now my go-to reference for this topic.",
		"Excellent breakdown of complex concepts.",
		"Much appreciated! Very informative post.",
		"I learned something new today. Thanks!",
		"This could benefit from more detailed explanations.",
		"Awesome content! Shared with my team.",
		"Quick question: does this work with older versions?",
		"Very practical and actionable advice.",
		"I've bookmarked this for future reference.",
		"The diagrams really help clarify things.",
		"Interesting approach, never thought of it that way.",
		"Could use some more real-world examples.",
		"Perfect timing! I needed this for my project.",
		"Well-researched and well-written article.",
		"I have some concerns about the performance implications.",
		"This is a game-changer for my workflow!",
		"Clear and straightforward. Exactly what I needed.",
		"Have you considered alternative approaches?",
		"Incredibly useful information. Thank you for sharing!",
		"I think there might be a better way to handle this.",
		"Outstanding tutorial! Very professional.",
		"This helped me pass my technical interview!",
		"Great post! Would you mind if I translate this?",
		"Solid advice backed by good examples.",
	}
)

func Seed(store store.Storage, db *sql.DB) {

	n_users, n_posts, n_comments := 100, 250, 500
	s_users, s_posts, s_comments := 100, 250, 500

	ctx := context.Background()

	users := generateUsers(n_users)
	tx, _ := db.BeginTx(ctx, nil)
	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			fmt.Printf("failed to create user %s: %v\n", user.Username, err)
			s_users--
		}
	}
	tx.Commit()

	posts := generatePosts(n_posts, users)

	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			fmt.Printf("failed to create post %s: %v\n", post.Title, err)
			s_posts--
		}
	}

	comments := generateComments(n_comments, users, posts)

	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			fmt.Printf("failed to create comment: %v\n", err)
			s_comments--
		}
	}

	log.Printf("Seeded %d users, %d posts, %d comments", s_users, s_posts, s_comments)
}

func generateUsers(num int) []*store.User {

	users := make([]*store.User, num)

	for i := 0; i < num; i++ {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@example.com",
		}
	}
	return users
}

func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)

	for i := 0; i < num; i++ {
		user := users[rand.Intn(len(users))]

		n_tags := rand.Intn(5)
		postTags := make([]string, n_tags)
		for j := 0; j < n_tags; j++ {
			postTags[j] = tags[rand.Intn(len(tags))]
		}

		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			Tags:    postTags,
		}
	}

	return posts
}

func generateComments(num int, users []*store.User, posts []*store.Post) []*store.Comment {
	commentsList := make([]*store.Comment, num)

	for i := 0; i < num; i++ {
		commentsList[i] = &store.Comment{
			PostID:  posts[rand.Intn(len(posts))].ID,
			UserID:  users[rand.Intn(len(users))].ID,
			Content: comments[rand.Intn(len(comments))],
		}
	}

	return commentsList
}
