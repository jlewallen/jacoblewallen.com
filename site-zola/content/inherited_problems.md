+++
title = "Inherited Problems"
date = 2012-05-30T00:00:00-08:00
draft = false
authors = ["admin"]
[taxonomies]
tags = ["programming"]
+++

When developers first begin learning object-orientated programming inheritance usually stands out more than any other language feature or practice. It’s introduction can come in many flavors, my favorite of which is the shape hierarchy. An abstract base class, Shape, has many derived concrete classes like Triangle, Circle, and Square. Shape itself has methods for calculating area, or perimeter that the derived classes can provide implementations for [^1]. It’s one of many possible examples (another is the Animal hierarchy) that demonstrates how inheritance, coupled with other practices like polymorphism and encapsulation can create powerful models of behavior and data.

In my experience, the real value and purpose of inheritance begins to get cloudy after that first introduction. In fact, I personally have used inheritance incorrectly for most of my career. More often than not it’s used as a mechanism to reuse code. How many of us have written or seen classes like this:

* AbstractEntity
* AbstractController
* AbstractService
* AbstractRepository

I’ve written classes just like these, adding methods that can be overridden in derived classes (AbstractEntity.saved, etc..) and gone on my merry way, happy that I’m reusing code and providing extensible points in the code base, completely unaware that I’ve laid the foundation for a future problem. In fact, these can sometimes start as interfaces masquerading as abstract classes. When you apply inheritance to only reuse code, you’re not doing so in order to create a polymorphic hierarchy, where different instances in the hierarchy can be substituted for one another. In the shape hierarchy an instance of a Shape can be a Triangle or a Square and the two are interchangeable. A non-developer would agree the two are members of the same hierarchy. That is not always the case if you have an AbstractController or AbstractEntity who’s reason for existence is to provide a repository for reusable methods to derived classes. In other words, there’s no behavioral or classifiable relationship between the types in the hierarchy outside of them residing in the same layer or functional area of the application (Controllers, Domain, Services), there’s only an implementation-level relationship.

Now there’s a coupling nightmare in the works as code living in the base class begins to grow, instantly bestowing behavior to all derived classes that many times they need to augment because the behavior doesn’t apply or the logic for applying that behavior is subtly different. So more abstract classes are introduced in the middle of the hierarchy (AbstractPersonEntity, AbstractTaggedEntity. AbstractAdminController), so that we can keep things DRY. What happens when a particular derived class needs to reuse code from two of these inner abstract classes, say a taggable person? Nope, many modern OO languages disallow multiple inheritance [^2]. Suddenly inheritance isn’t paying off at all.

Inheritance is one of the most abused patterns in object orientated programming. Many of us have used inheritance as a mechanism to keep things DRY and reuse code. Odds are, it was the wrong choice. What to do? Composition. If a particular behavior is common and needs to be reused, create a class that can be used. These classes always have a well defined interface and can have multiple implementations the rest of the system can take advantage of. A controller for editing people can use a repository that performs heavy auditing, an entity that can be tagged can defer to the reusable tagging code lying in a TagSet class. As this is applied throughout the code it becomes clearer how certain implementations are solving common problems, and if changes to one particular part of the system need to occur, they can be made in isolation much easier. Code re-used via composition benefits from a much narrower contract than large, complicated abstract base classes and it’s much easier to test and make changes to than if it were sitting among a dozen other concerns.

* http://en.wikipedia.org/wiki/Inheritance_(computer_science)
* http://en.wikipedia.org/wiki/Liskov_substitution_principle
* http://en.wikipedia.org/wiki/Composition_over_inheritance
* http://en.wikipedia.org/wiki/Separation_of_concerns

[^1]: My simplified shape example is also kind of awkward in that Shape as described isn’t providing behavior, though the example could be expanded upon so that it does in a reasonable way.
[^2]: Not to imply that I consider multiple inheritance a good solution.
